package raft

import (
	"sync"
	"errors"
	"os"
	"encoding/json"
	"path"
	"io/ioutil"
	"time"
)

const (
	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultTmpConfig = "config.tmp"
	DefaultLog = "log"
	DefaultSnapshot = "snapshot"
)

type Server struct {
	mu 								sync.RWMutex
	waitGroup 						sync.WaitGroup

	address							string
	leader							string

	data_dir    					string
	rpcs							*RPCs
	peers      						map[string]*Node
	nodesMut 						sync.RWMutex
	detectTicker					*time.Ticker

	state							State
	currentTerm						uint64
	votedFor   						string
	log								[]*Entry

	commitIndex						uint64
	lastApplied						uint64
	lastRPCTime						time.Time

	//candidate
	votes	 						chan *Vote
	voteDic	 						map[string]int
	voteTotal	 					int
	voteCount	 					int
	quorum							int

	//leader
	nextIndex						map[string]uint64
	matchIndex						map[string]uint64

	stop							chan bool
	running							bool

	changeState 					chan int
	ticker							*time.Ticker

	hearbeatTick					time.Duration
	waitHearbeatTimeout				time.Duration
	electionTimeout					time.Duration

}

func NewServer(address,data_dir string)(*Server,error){
	if address == "" {
		return nil, errors.New("address is empty")
	}
	if data_dir == "" {
		data_dir=DefaultDataDir
	}
	mkdir(data_dir)
	s := &Server{
		address:address,
		data_dir:data_dir,
		rpcs:newRPCs(address,[]string{}),
		peers:make(map[string]*Node),
		detectTicker:time.NewTicker(DefaultDetectTick),
		stop:make(chan bool),
		changeState: make(chan int,1),
		votes: make(chan *Vote,1),
		voteDic:make(map[string]int),
		ticker:time.NewTicker(time.Millisecond),
		waitHearbeatTimeout:DefaultWaitHearbeatTimeout,
		electionTimeout:DefaultElectionTimeout,
		hearbeatTick:DefaultHearbeatTick,
	}
	s.state=newFollowerState(s)
	listenAndServe(address,s)
	go s.run()
	return s,nil
}
func (s *Server) run() {
	for{
		select {
		case <-s.stop:
			goto endfor
		case v:=<-s.votes:
			if _,ok:=s.voteDic[v.Key()];ok{
				return
			}
			s.voteDic[v.Key()]=v.vote
			if v.term==s.currentTerm+1{
				s.voteTotal+=1
				s.voteCount+=v.vote
			}
		case <-s.detectTicker.C:
			s.detectNodes()
		case <-s.ticker.C:
			select {
			case i := <-s.changeState:
				if i == 1 {
					s.SetState(s.state.NextState())
				} else if i == -1 {
					s.SetState(s.state.PreState())
				}else if i == 0 {
					s.state.Init()
					return
				}
			default:
				s.state.Update()
			}

		}
	}
endfor:
}

func (s *Server) resetLastRPCTime() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastRPCTime = time.Now()
}

func (s *Server) SetState(state State) {
	s.state = state
}

func (s *Server) GetState()State {
	return s.state
}
func (s *Server) ChangeState(i int){
	s.changeState<-i
}
func (s *Server) State()string {
	return s.state.String()
}
func (s *Server) Term()uint64 {
	return s.currentTerm
}
func (s *Server) Leader()string {
	return s.leader
}
func (s *Server) isLeader()bool {
	if s.leader==s.address{
		return true
	}else{
		return false
	}
}
func (s *Server) AddNode(address string) error {
	s.nodesMut.Lock()
	defer s.nodesMut.Unlock()
	if s.peers[address] != nil {
		return nil
	}
	if s.address != address {
		node := newNode(s,address)
		s.peers[address] = node
	}

	s.saveConfig()
	return nil
}
func (s *Server) NodesLen() int {
	s.nodesMut.RLock()
	defer s.nodesMut.RUnlock()
	return len(s.peers)+1
}
func (s *Server) Quorum() int {
	s.nodesMut.RLock()
	defer s.nodesMut.RUnlock()
	return (len(s.peers)+1)/2+1
}
func (s *Server) requestVotes() error {
	s.nodesMut.RLock()
	defer s.nodesMut.RUnlock()
	if len(s.votes)>0{
		<-s.votes
	}
	s.voteDic=make(map[string]int)
	s.voteCount=0
	s.voteTotal=0
	s.votes<-&Vote{candidateId:s.address,vote:1,term:s.currentTerm+1}
	for _,v :=range s.peers{
		if v.alive==true{
			go v.requestVote()
		}
	}
	return nil
}
func (s *Server) detectNodes() error {
	s.nodesMut.RLock()
	defer s.nodesMut.RUnlock()
	for _,v :=range s.peers{
		if v.alive==false{
			go v.ping()
		}
	}
	return nil
}
func (s *Server) heartbeats() error {
	s.nodesMut.RLock()
	defer s.nodesMut.RUnlock()
	for _,v :=range s.peers{
		if v.alive==true{
			go v.heartbeat()
		}
	}
	return nil
}

func (s *Server) saveConfig() {
	peers := make([]string, len(s.peers))
	i := 0
	for _, node := range s.peers {
		peers[i] = node.address
		i++
	}
	c := &Configuration{
		Node:s.address,
		Peers:       peers,
	}
	b, _ := json.Marshal(c)
	config_path := path.Join(s.data_dir, DefaultConfig)
	tmp_config_path := path.Join(s.data_dir, DefaultTmpConfig)
	err := writeToFile(tmp_config_path, b)
	if err != nil {
		panic(err)
	}
	os.Rename(tmp_config_path, config_path)
}

func (s *Server) loadConfig() error {
	config_path := path.Join(s.data_dir, DefaultConfig)
	b, err := ioutil.ReadFile(config_path)
	if err != nil {
		return nil
	}
	config := &Configuration{}
	if err = json.Unmarshal(b, config); err != nil {
		return err
	}
	//s.log.updateCommitIndex(config.CommitIndex)
	return nil
}