package raft

import (
	"sync"
	"errors"
	"encoding/json"
	"time"
)

const (
	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultTmpConfig = "config.tmp"
	DefaultTmp = ".tmp"
	DefaultLog = "log"
	DefaultSnapshot = "snapshot"
)

type Node struct {
	mu 								sync.RWMutex
	nodesMut 						sync.RWMutex
	waitGroup 						sync.WaitGroup
	onceStart 						sync.Once
	onceStop 						sync.Once

	running							bool
	stop							chan bool
	stoped							bool
	enabled							bool

	address							string
	leader							string

	//config
	hearbeatTick					time.Duration
	normalOperationTimeout			time.Duration
	electionTimeout					time.Duration
	storage *Storage
	raft 							Raft
	rpcs							*RPCs
	server							*Server
	clients      					map[string]*Client
	detectTicker					*time.Ticker


	state							State
	changeStateChan 				chan int
	ticker							*time.Ticker

	//persistent state on all servers
	currentTerm						uint64
	votedFor   						string
	log								[]*Entry

	//volatile state on all servers
	commitIndex						uint64
	lastApplied						uint64


	//volatile state on leader
	nextIndex						map[string]uint64
	matchIndex						map[string]uint64


	lastRPCTime						time.Time
	lastLogIndex					uint64
	lastLogTerm						uint64

	//candidate
	votes	 						*Votes

}

func NewNode(address,data_dir string)(*Node,error){
	if address == "" {
		return nil, errors.New("address can not be empty")
	}
	if data_dir == "" {
		data_dir=DefaultDataDir
	}
	n := &Node{
		address:address,
		storage:newStorage(data_dir),
		rpcs:newRPCs(address,[]string{}),

		clients:make(map[string]*Client),
		detectTicker:time.NewTicker(DefaultDetectTick),
		stop:make(chan bool,1),
		changeStateChan: make(chan int,1),
		ticker:time.NewTicker(DefaultNodeTick),
		normalOperationTimeout:DefaultNormalOperationTimeout,
		electionTimeout:DefaultElectionTimeout,
		hearbeatTick:DefaultHearbeatTick,
	}
	n.state=newFollowerState(n)
	n.raft=newRaft(n)
	n.server=newServer(address,n)
	n.votes=newVotes(n)
	return n,nil
}
func (n *Node) run() {
	n.running=true
	for{
		select {
		case <-n.stop:
			goto endfor
		case v:=<-n.votes.vote:
			n.votes.AddVote(v)
		case <-n.detectTicker.C:
			n.detectNodes()
		case <-n.ticker.C:
			select {
			case i := <-n.changeStateChan:
				if i == 1 {
					n.setState(n.state.NextState())
				} else if i == -1 {
					n.setState(n.state.StepDown())
				}else if i == 0 {
					n.state.Init()
				}
			default:
				n.state.Update()
			}

		}
	}
endfor:
}
func (n *Node) Start() {
	n.onceStart.Do(func() {
		n.server.listenAndServe()
		go n.run()
	})
}
func (n *Node)Running() bool{
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}
func (n *Node) Stop() {
	n.onceStop.Do(func() {
		n.stop<-true
		n.stoped=true
	})
}
func (n *Node)Stoped() bool{
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stoped
}
func (n *Node)Enabled() bool{
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.enabled
}

func (n *Node) LastRPCTime()time.Time{
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lastRPCTime
}
func (n *Node) resetLastRPCTime() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastRPCTime = time.Now()
}

func (n *Node) setState(state State) {
	n.state = state
}

func (n *Node) GetState()State {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state
}

func (n *Node) stepDown(){
	n.changeState(-1)
}
func (n *Node) nextState(){
	n.changeState(1)
}

func (n *Node) stay(){
	n.changeState(0)
}

func (n *Node) changeState(i int){
	n.changeStateChan<-i
}

func (n *Node) State()string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state.String()
}
func (n *Node) Term()uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.currentTerm
}
func (n *Node) Leader()string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leader
}
func (n *Node) IsLeader()bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.leader==n.address{
		return true
	}else{
		return false
	}
}
func (n *Node) AddNode(address string) error {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	if n.clients[address] != nil {
		return nil
	}
	if n.address != address {
		client := newClient(n,address)
		n.clients[address] = client
	}
	n.votes.Reset(len(n.clients)+1)
	n.saveConfig()
	return nil
}
func (n *Node) NodesCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return len(n.clients)+1
}
func (n *Node) Quorum() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return (len(n.clients)+1)/2+1
}
func (n *Node) AliveCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	cnt:=1
	for _,v:=range n.clients{
		if v.alive==true{
			cnt+=1
		}
	}
	return cnt
}
func (n *Node) FollowerCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	cnt:=1
	for _,v:=range n.clients{
		if v.follower==true{
			cnt+=1
		}
	}
	return cnt
}
func (n *Node) requestVotes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	n.votes.Clear()
	n.votes.vote<-&Vote{candidateId:n.address,vote:1,term:n.currentTerm}
	for _,v :=range n.clients{
		if v.alive==true{
			go v.requestVote()
		}
	}
	return nil
}
func (n *Node) detectNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.clients{
		if v.alive==false{
			go v.ping()
		}
	}
	return nil
}
func (n *Node) heartbeats() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.clients{
		if v.alive==true{
			go v.heartbeat()
		}
	}
	return nil
}

func (n *Node) saveConfig() {
	n.mu.Lock()
	defer n.mu.Unlock()
	clients := make([]string, len(n.clients))
	i := 0
	for _, node := range n.clients {
		clients[i] = node.address
		i++
	}
	c := &Configuration{
		Name:n.address,
		Others:clients,
	}
	b, _ := json.Marshal(c)
	n.storage.Persistence(DefaultConfig,b)
}

func (n *Node) loadConfig() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	b, err := n.storage.Load(DefaultConfig)
	if err != nil {
		return nil
	}
	config := &Configuration{}
	if err = json.Unmarshal(b, config); err != nil {
		return err
	}
	//n.log.updateCommitIndex(config.CommitIndex)
	return nil
}