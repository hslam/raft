package raft

import (
	"sync"
	"fmt"
	"time"
	"hslam.com/mgit/Mort/timer"
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

	host    						string
	port 							int
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
	peers      						map[string]*Peer
	detectTicker					*timer.Ticker

	keepAliveTicker					*timer.Ticker

	state							State
	changeStateChan 				chan int
	ticker							*timer.Ticker

	//persistent state on all servers
	currentTerm						*PersistentUint64
	votedFor   						*PersistentString

	//volatile state on all servers
	commitIndex						*PersistentUint64



	lastLogIndex					*PersistentUint64
	lastLogTerm						uint64
	nextIndex						uint64

	//init
	recoverLogIndex					uint64
	lastPrintLastLogIndex			uint64
	lastPrintCommitIndex			uint64
	lastPrintLastApplied			uint64

	//candidate
	votes	 						*Votes
	election						*Election

	raftCodec						Codec
	codec							Codec

	context 						interface{}
	commandType						*CommandType
	pipeline						*Pipeline
	pipelineChan					chan bool


}

func NewNode(host string, port int,data_dir string,context interface{})(*Node,error){
	if data_dir == "" {
		data_dir=DefaultDataDir
	}
	address:=fmt.Sprintf("%s:%d",host,port)
	n := &Node{
		host:host,
		port:port,
		address:address,
		storage:newStorage(data_dir),
		rpcs:newRPCs([]string{}),
		peers:make(map[string]*Peer),
		detectTicker:timer.NewFuncTicker(DefaultDetectTick),
		keepAliveTicker:timer.NewFuncTicker(DefaultKeepAliveTick),
		ticker:timer.NewFuncTicker(DefaultNodeTick),
		stop:make(chan bool,1),
		changeStateChan: make(chan int,1),
		hearbeatTick:DefaultHearbeatTick,
		raftCodec:new(ProtoCodec),
		codec:new(JsonCodec),
		context:context,
		commandType:&CommandType{types:make(map[int32]Command)},
		nextIndex:1,
	}
	n.votes=newVotes(n)
	n.stateMachine=newStateMachine(n)
	n.configuration=newConfiguration(n)
	n.log=newLog(n)
	n.election=newElection(n,DefaultElectionTimeout)
	n.raft=newRaft(n)
	n.server=newServer(n,fmt.Sprintf(":%d",port))
	n.currentTerm=newPersistentUint64(n,DefaultTerm)
	n.lastLogIndex=newPersistentUint64(n,DefaultLastLogIndex)
	n.commitIndex=newPersistentUint64(n,DefaultCommitIndex)
	n.votedFor=newPersistentString(n,DefaultVoteFor)
	n.state=newFollowerState(n)
	n.pipeline=NewPipeline(n,DefaultMaxConcurrency)
	n.pipelineChan = make(chan bool,DefaultMaxConcurrency)

	return n,nil
}
func (n *Node) Start() {
	n.log.recover()
	n.election.Reset()
	n.onceStart.Do(func() {
		n.server.listenAndServe()
		go n.run()
	})
	n.running=true
}
func (n *Node) run() {
	timer.NewFuncTicker(DefaultNodeTracePrintTick, func() {
		n.print()
	})
	n.ticker.Tick(func() {
		if !n.running{
			return
		}
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
	})
	n.detectTicker.Tick(func() {
		n.detectNodes()
	})
	n.keepAliveTicker.Tick(func() {
		n.keepAliveNodes()
	})
	for{
		select {
		case <-n.stop:
			goto endfor
		case v:=<-n.votes.vote:
			n.votes.AddVote(v)
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
	if n.state==nil{
		return ""
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state.String()
}

func (n *Node) Term()uint64 {
	if !n.running{
		return 0
	}
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
	return n.isLeader()
}
func (n *Node) isLeader()bool {
	if !n.running{
		return false
	}
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
func (n *Node)SetSnapshot(snapshot Snapshot){
	n.stateMachine.SetSnapshot(snapshot)
}
func (n *Node)RegisterCommand(command Command) (error){
	if command == nil {
		return ErrCommandNil
	}else if command.Type() < 0 {
		return ErrCommandTypeMinus
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.commandType.register(command)
}
func (n *Node)Do(command Command) (interface{},error){
	if !n.running{
		return nil,ErrNotRunning
	}
	if command == nil {
		return nil,ErrCommandNil
	}else if command.Type() < 0 {
		return nil,ErrCommandTypeMinus
	}
	n.pipelineChan<-true
	var reply interface{}
	var err error
	if n.IsLeader(){
		if n.commandType.exists(command){
			invoker:=newInvoker(command,false,n.codec)

			replyChan:=make(chan interface{},1)
			errChan:=make(chan error,1)
			ch:=NewPipelineCommand(invoker,replyChan,errChan)
			n.pipeline.pipelineCommandChan<-ch
			select {
			case reply=<-replyChan:
			case err=<-errChan:
			case <-time.After(DefaultCommandTimeout):
				err=ErrCommandTimeout
			}
			close(replyChan)
			close(errChan)
		}else {
			err=ErrCommandNotRegistered
		}
	}else {
		err=ErrNotLeader
	}

	<-n.pipelineChan
	return reply,err
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
	return n.quorum()
}
func (n *Node) quorum() int {
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


func (n *Node) requestVotes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	n.votes.Clear()
	n.votes.vote<-newVote(n.address,n.currentTerm.Id(),1)
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
func (n *Node) keepAliveNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		go v.ping()
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
func (n *Node) check() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if v.alive==true{
			go v.check()
		}
	}
	return nil
}

func (n *Node) printPeers(){
	if !n.IsLeader(){
		return
	}
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		Tracef("Node.printPeers %s %d",v.address,v.nextIndex)
	}
}

func (n *Node) print(){
	if n.lastLogIndex.Id()>n.lastPrintLastLogIndex{
		Tracef("Node.print %s lastLogIndex %d==>%d",n.address,n.lastPrintLastLogIndex,n.lastLogIndex.Id())
		n.lastPrintLastLogIndex=n.lastLogIndex.Id()
	}
	if n.commitIndex.Id()>n.lastPrintCommitIndex{
		Tracef("Node.print %s commitIndex %d==>%d",n.address,n.lastPrintCommitIndex,n.commitIndex.Id())
		n.lastPrintCommitIndex=n.commitIndex.Id()
	}
	if n.stateMachine.lastApplied>n.lastPrintLastApplied{
		Tracef("Node.print %s lastApplied %d==>%d",n.address,n.lastPrintLastApplied,n.stateMachine.lastApplied)
		n.lastPrintLastApplied=n.stateMachine.lastApplied
	}
	//n.printPeers()
}
func (n *Node) Commit() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	//var commitIndex=n.commitIndex
	if len(n.peers)==0{
		index:=n.lastLogIndex.Id()
		if index>n.commitIndex.Id(){
			n.commitIndex.Set(index)
			//Tracef("Node.Commit %s commitIndex %d==>%d",n.address,commitIndex,n.commitIndex)
		}
		if n.commitIndex.Id()<=n.recoverLogIndex&&n.commitIndex.Id()>n.stateMachine.lastApplied{
			//var lastApplied=n.stateMachine.lastApplied
			n.log.applyCommited()
			//Tracef("Node.Commit %s lastApplied %d==>%d",n.address,lastApplied,n.stateMachine.lastApplied)
		}
		return nil
	}
	quorum:=n.quorum()
	var lastLogIndexCount =make(map[uint64]int)
	var lastLogIndexs	=make([]uint64,0)
	for _,v :=range n.peers{
		if _,ok:=lastLogIndexCount[v.nextIndex-1];!ok{
			lastLogIndexCount[v.nextIndex-1]=1
			lastLogIndexs=append(lastLogIndexs, v.nextIndex-1)
		}else {
			lastLogIndexCount[v.nextIndex-1]+=1
		}
	}
	quickSort(lastLogIndexs,-999,-999)
	for i:=0;i<len(lastLogIndexs);i++{
		index:=lastLogIndexs[i]
		if v,ok:=lastLogIndexCount[index];ok{
			if v+1>=quorum{
				if index>n.commitIndex.Id(){
					n.commitIndex.Set(index)
					//Tracef("Node.Commit %s commitIndex %d==>%d",n.address,commitIndex,n.commitIndex)
					if n.commitIndex.Id()<=n.recoverLogIndex{
						//var lastApplied=n.stateMachine.lastApplied
						n.log.applyCommited()
						//Tracef("Node.Commit %s lastApplied %d==>%d",n.address,lastApplied,n.stateMachine.lastApplied)
					}
				}
				break
			}
		}
	}

	return nil
}