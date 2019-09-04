package raft

import (
	"sync"
	"errors"
	"time"
)

const (
	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultLog = "log"
	DefaultTerm = "term"
	DefaultSnapshot = "snapshot"
	DefaultTmp = ".tmp"
)

type Node struct {
	mu 								sync.RWMutex
	nodesMut 						sync.RWMutex
	syncMut 						sync.RWMutex
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
	storage 						*Storage
	raft 							Raft
	rpcs							*RPCs
	server							*Server
	configuration 					*Configuration
	log								*Log

	stateMachine					*StateMachine
	snapshot						*Snapshot
	peers      						map[string]*Peer
	detectTicker					*time.Ticker


	state							State
	changeStateChan 				chan int
	ticker							*time.Ticker

	//persistent state on all servers
	currentTerm						*Term
	votedFor   						string

	//volatile state on all servers
	commitIndex						uint64


	//volatile state on leader
	//nextIndex						map[string]uint64
	//matchIndex					map[string]uint64

	logIndex						uint64


	lastLogIndex					uint64
	lastLogTerm						uint64

	//follower
	prevLogIndex 					uint64
	prevLogTerm						uint64


	//candidate
	votes	 						*Votes
	election						*Election

	raftCodec						Codec
	codec							Codec

	context 						interface{}
	commandType						*CommandType
	pipeline						*Pipeline
}

func NewNode(address,data_dir string,context interface{})(*Node,error){
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
		peers:make(map[string]*Peer),
		detectTicker:time.NewTicker(DefaultDetectTick),
		stop:make(chan bool,1),
		changeStateChan: make(chan int,1),
		ticker:time.NewTicker(DefaultNodeTick),
		hearbeatTick:DefaultHearbeatTick,
		raftCodec:new(ProtoCodec),
		codec:new(JsonCodec),
		context:context,
		commandType:&CommandType{make(map[uint32]Command)},
	}
	n.votes=newVotes(n)

	n.configuration=newConfiguration(n)
	n.configuration.load()
	n.log=newLog(n)
	n.election=newElection(n,DefaultElectionTimeout)
	n.raft=newRaft(n)
	n.server=newServer(n,address)
	n.currentTerm=newTerm(n)
	n.state=newFollowerState(n)
	n.stateMachine=newStateMachine(n)
	n.snapshot=newSnapshot(n)
	n.pipeline=NewPipeline(n,DefaultMaxConcurrency)

	return n,nil
}
func (n *Node) Start() {
	n.onceStart.Do(func() {
		n.server.listenAndServe()
		go n.run()
	})
	n.log.recovery()
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
					n.state.Reset()
				}
			default:
				n.state.Update()
			}

		}
	}
endfor:
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
	return n.currentTerm.Id()
}
func (n *Node) Leader()string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leader
}
func (n *Node) IsLeader()bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.leader==n.address&&n.state.String()==Leader{
		return true
	}else{
		return false
	}
}

func (n *Node) SetCodec(codec Codec){
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.codec=codec
}
func (n *Node) SetContext(context interface{}){
	n.mu.Lock()
	defer n.mu.Unlock()
	n.context=context
}

func (n *Node) Context()interface{}{
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.context
}
func (n *Node)Register(command Command) (error){
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.commandType.register(command)
}
func (n *Node)Do(command Command) (interface{},error){
	if n.IsLeader(){
		if !n.commandType.exists(command){
			return nil,ErrCommandNotRegistered
		}
		invoker:=newInvoker(command,false,n.codec)
		ch:=NewPipelineCommand(invoker)
		n.pipeline.pipelineCommandChan<-ch
		select {
		case res:=<-ch.cbChan:
			return res,nil
		case err:=<-ch.cbErrorChan:
			return nil,err
		case <-time.After(DefaultCommandTimeout):
			return nil,ErrCommandTimeout
		}
	}
	return nil,ErrNotLeader
}

func (n *Node) SetNode(addrs []string) error {
	if !n.configuration.isPeersChanged(addrs){
		return nil
	}
	n.peers=make(map[string]*Peer)
	for _,address:=range addrs{
		n.addNode(address)
	}
	n.configuration.SetPeers(addrs)
	n.configuration.save()
	return nil
}
func (n *Node) addNode(address string) error {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	if n.peers[address] != nil {
		return nil
	}
	if n.address != address {
		client := newPeer(n,address)
		n.peers[address] = client
	}
	n.votes.Reset(len(n.peers)+1)
	return nil
}
func (n *Node) NodesCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return len(n.peers)+1
}
func (n *Node) Quorum() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return (len(n.peers)+1)/2+1
}
func (n *Node) AliveCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	cnt:=1
	for _,v:=range n.peers{
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
	for _,v:=range n.peers{
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
	n.votes.vote<-&Vote{candidateId:n.address,vote:1,term:n.currentTerm.Id()}
	for _,v :=range n.peers{
		if v.alive==true{
			go v.requestVote()
		}
	}
	return nil
}
func (n *Node) detectNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if v.alive==false{
			go v.ping()
		}
	}
	return nil
}
func (n *Node) heartbeats() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if v.alive==true{
			go v.heartbeat()
		}
	}
	return nil
}

func (n *Node) appendEntries(entries []*Entry) error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,p :=range n.peers{
		go p.appendEntries(entries)
	}
	return nil
}
