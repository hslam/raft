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
	heartbeatTick					time.Duration
	storage 						*Storage
	raft 							Raft
	rpcs							*RPCs
	server							*Server
	configuration 					*Configuration
	log								*Log

	stateMachine					*StateMachine
	peers      						map[string]*Peer
	detectTicker					*time.Ticker
	keepAliveTicker					*time.Ticker
	printTicker						*time.Ticker
	checkLogTicker					*time.Ticker
	state							State
	changeStateChan 				chan int
	ticker							*timer.FuncTicker

	//persistent state on all servers
	currentTerm						*PersistentUint64
	votedFor   						*PersistentString

	//volatile state on all servers
	commitIndex						*PersistentUint64

	lastLogIndex					uint64
	//lastLogIndex					*PersistentUint64
	lastLogTerm						uint64
	nextIndex						uint64

	//init
	recoverLogIndex					uint64
	lastPrintLastLogIndex			uint64
	lastPrintCommitIndex			uint64
	lastPrintLastApplied			uint64
	lastPrintNextIndex				uint64

	//candidate
	votes	 						*Votes
	election						*Election

	//leader
	lease 							bool
	ready 							bool

	raftCodec						Codec
	commandCodec					Codec
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
		rpcs:newRPCs([]string{}),
		peers:make(map[string]*Peer),
		printTicker:time.NewTicker(DefaultNodeTracePrintTick),
		detectTicker:time.NewTicker(DefaultDetectTick),
		keepAliveTicker:time.NewTicker(DefaultKeepAliveTick),
		checkLogTicker:time.NewTicker(DefaultCheckLogTick),
		ticker:timer.NewFuncTicker(DefaultNodeTick,nil),
		stop:make(chan bool,1),
		changeStateChan: make(chan int,1),
		heartbeatTick:DefaultHeartbeatTick,
		raftCodec:new(ProtoCodec),
		commandCodec:new(JsonCodec),
		codec:new(JsonCodec),
		context:context,
		commandType:&CommandType{types:make(map[int32]Command)},
		nextIndex:1,
	}
	n.storage=newStorage(n,data_dir)
	n.votes=newVotes(n)
	n.stateMachine=newStateMachine(n)
	n.configuration=newConfiguration(n)
	n.log=newLog(n)
	n.election=newElection(n,DefaultElectionTimeout)
	n.raft=newRaft(n)
	n.server=newServer(n,fmt.Sprintf(":%d",port))
	n.currentTerm=newPersistentUint64(n,DefaultTerm)
	//n.lastLogIndex=newPersistentUint64(n,DefaultLastLogIndex)
	n.commitIndex=newPersistentUint64(n,DefaultCommitIndex)
	n.votedFor=newPersistentString(n,DefaultVoteFor)
	n.state=newFollowerState(n)
	n.pipeline=NewPipeline(n,DefaultMaxConcurrency)
	n.pipelineChan = make(chan bool,DefaultMaxConcurrency)
	n.registerCommand(&noOperationCommand{})
	return n,nil
}
func (n *Node) Start() {
	n.onceStart.Do(func() {
		n.recover()
		n.server.listenAndServe()
		go n.run()
	})
	n.election.Reset()
	n.running=true
}
func (n *Node) run() {
	n.ticker.Tick(func() {
		if !n.running{
			return
		}
		func(){
			defer func() {if err := recover(); err != nil {}}()
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
		}()
	})
	for {
		select {
		case <-n.printTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				n.print()
			}()
		case <-n.keepAliveTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				n.keepAliveNodes()
			}()
		case <-n.detectTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				n.detectNodes()
			}()
		case <-n.checkLogTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				n.checkLog()
			}()
		case v:=<-n.votes.vote:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				n.votes.AddVote(v)
			}()
		case <-n.stop:
			goto endfor
		}
	}
endfor:
	close(n.stop)
	n.printTicker.Stop()
	n.printTicker=nil
	n.keepAliveTicker.Stop()
	n.keepAliveTicker=nil
	n.detectTicker.Stop()
	n.detectTicker=nil
	n.ticker.Stop()
	n.ticker=nil
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
	if len(n.changeStateChan)==1{
		return
	}
	n.changeStateChan<-i
}

func (n *Node) State()string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.state==nil{
		return ""
	}
	return n.state.String()
}

func (n *Node) Term()uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.term()
}
func (n *Node) term()uint64 {
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
func (n *Node) Ready()bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.ready
}
func (n *Node) Lease()bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lease
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
	if n.leader==n.address&&n.state.String()==Leader&&n.lease{
		return true
	}else{
		return false
	}
}
func (n *Node) Address()string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.address
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
func (n *Node)SetSnapshotSyncType(snapshotSyncType SnapshotSyncType){
	n.stateMachine.SetSnapshotSyncType(snapshotSyncType)
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
func (n *Node)registerCommand(command Command) (error){
	if command == nil {
		return ErrCommandNil
	}
	return n.commandType.register(command)
}
func (n *Node)Do(command Command) (interface{},error){
	if command.Type() < 0 {
		return nil,ErrCommandTypeMinus
	}
	return n.do(command,DefaultCommandTimeout)
}
func (n *Node)do(command Command,timeout time.Duration) (interface{},error){
	if !n.running{
		return nil,ErrNotRunning
	}
	if command == nil {
		return nil,ErrCommandNil
	}
	n.pipelineChan<-true
	var reply interface{}
	var err error
	if n.IsLeader(){
		if n.commandType.exists(command){
			var invoker RaftCommand
			if command.Type()>=0{
				invoker=newInvoker(command,false,n.codec)
			}else {
				invoker=newInvoker(command,false,n.commandCodec)
			}
			replyChan:=make(chan interface{},1)
			errChan:=make(chan error,1)
			ch:=NewPipelineCommand(invoker,replyChan,errChan)
			n.pipeline.pipelineCommandChan<-ch
			select {
			case reply=<-replyChan:
			case err=<-errChan:
			case <-time.After(timeout):
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
func (n *Node) fast() bool {
	return n.isLeader()&&len(n.peers)>0
}
func (n *Node) Peers() []string {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	peers:=make([]string,0,len(n.peers))
	for _,v:=range n.peers{
		peers=append(peers,v.address)
	}
	return peers
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
	return len(n.peers)/2+1
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
		if v.alive{
			go v.heartbeat()
		}
	}
	return nil
}
func (n *Node) install() bool {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if !v.install{
			return false
		}
	}
	return true
}
func (n *Node) check() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if v.alive==true{
			v.check()
		}
	}
	return nil
}
func (n *Node) reset(){
	n.recoverLogIndex=0
	n.lastPrintNextIndex=1
	n.lastPrintLastApplied=0
	n.lastPrintLastLogIndex=0
	n.lastPrintCommitIndex=0
	n.nextIndex=1
	n.lastLogIndex=0
	n.lastLogTerm=0
	n.commitIndex.Set(0)
	n.stateMachine.lastApplied=0
	n.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(0)
	n.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(0)
	n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
}
func (n *Node) load(){
	n.log.load()
	n.stateMachine.load()
}
func (n *Node) recover() error {
	Tracef("Node.recover %s start",n.address)
	if n.storage.IsEmpty(DefaultIndex){
		if !n.storage.IsEmpty(DefaultTarGz){
			n.stateMachine.snapshotReadWriter.untar()
		}
	}else if n.storage.IsEmpty(DefaultLog){
		if !n.storage.IsEmpty(DefaultTarGz){
			n.stateMachine.snapshotReadWriter.untar()
		}
	}
	n.load()
	recoverApplyTicker:=time.NewTicker(time.Second)
	recoverApplyStop:=make(chan bool,1)
	go func() {
		for{
			select {
			case <-recoverApplyTicker.C:
				n.print()
			case <-recoverApplyStop:
				goto endfor
			}
		}
	endfor:
	}()
	n.stateMachine.recover()
	n.log.applyCommited()
	n.print()
	recoverApplyStop<-true
	Tracef("Node.recover %s finish",n.address)
	return nil
}
func (n *Node) checkLog() error {
	if n.storage.IsEmpty(DefaultCommitIndex){
		n.commitIndex.save()
	}
	if n.storage.IsEmpty(DefaultLastIncludedIndex){
		n.stateMachine.snapshotReadWriter.lastIncludedIndex.save()
	}
	if n.storage.IsEmpty(DefaultLastIncludedTerm){
		n.stateMachine.snapshotReadWriter.lastIncludedTerm.save()
	}
	if n.storage.IsEmpty(DefaultTerm){
		n.currentTerm.save()
	}
	if n.storage.IsEmpty(DefaultConfig){
		n.configuration.save()
	}
	if n.storage.IsEmpty(DefaultVoteFor){
		n.votedFor.save()
	}
	if n.storage.IsEmpty(DefaultSnapshot)&&n.stateMachine.snapshot!=nil&&!n.storage.IsEmpty(DefaultIndex)&&!n.storage.IsEmpty(DefaultLog)&&!n.storage.IsEmpty(DefaultCommitIndex){
		n.stateMachine.SaveSnapshot()
	}
	if n.isLeader()&&n.storage.IsEmpty(DefaultTarGz)&&!n.storage.IsEmpty(DefaultIndex)&&!n.storage.IsEmpty(DefaultLog)&&!n.storage.IsEmpty(DefaultCommitIndex)&&!n.storage.IsEmpty(DefaultSnapshot){
		n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
		n.stateMachine.snapshotReadWriter.Tar()
	}
	if n.storage.IsEmpty(DefaultIndex){
		if !n.storage.IsEmpty(DefaultTarGz){
			n.stateMachine.snapshotReadWriter.untar()
			n.load()
		}else {
			n.nextIndex=1
			//n.reset()
		}
	}else if n.storage.IsEmpty(DefaultLog){
		if !n.storage.IsEmpty(DefaultTarGz){
			n.stateMachine.snapshotReadWriter.untar()
			n.load()
		}else {
			n.nextIndex=1
			//n.reset()
		}
	}
	return nil
}

func (n *Node) printPeers(){
	if !n.isLeader(){
		return
	}
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _,v :=range n.peers{
		if v.nextIndex>v.lastPrintNextIndex{
			Tracef("Node.printPeers %s nextIndex %d==>%d",v.address,v.lastPrintNextIndex,v.nextIndex)
			v.lastPrintNextIndex=v.nextIndex
		}
	}
}

func (n *Node) print(){
	if n.lastLogIndex>n.lastPrintLastLogIndex{
		Tracef("Node.print %s lastLogIndex %d==>%d",n.address,n.lastPrintLastLogIndex,n.lastLogIndex)
		n.lastPrintLastLogIndex=n.lastLogIndex
	}
	if n.commitIndex.Id()>n.lastPrintCommitIndex{
		Tracef("Node.print %s commitIndex %d==>%d",n.address,n.lastPrintCommitIndex,n.commitIndex.Id())
		n.lastPrintCommitIndex=n.commitIndex.Id()
	}
	if n.stateMachine.lastApplied>n.lastPrintLastApplied{
		Tracef("Node.print %s lastApplied %d==>%d",n.address,n.lastPrintLastApplied,n.stateMachine.lastApplied)
		n.lastPrintLastApplied=n.stateMachine.lastApplied
	}
	if n.nextIndex>n.lastPrintNextIndex{
		Tracef("Node.print %s nextIndex %d==>%d",n.address,n.lastPrintNextIndex,n.nextIndex)
		n.lastPrintNextIndex=n.nextIndex
	}
	n.printPeers()
}
func (n *Node) Commit() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	//var commitIndex=n.commitIndex
	if len(n.peers)==0{
		index:=n.lastLogIndex
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
	for i:=len(lastLogIndexs)-1;i>=0;i--{
		index:=lastLogIndexs[i]
		if v,ok:=lastLogIndexCount[index];ok{
			if v+1>=quorum{
				if index>n.commitIndex.Id()&&index<n.nextIndex{
					n.commitIndex.Set(index)
					//Tracef("Node.Commit %s commitIndex %d==>%d",n.address,commitIndex,n.commitIndex)
				}
				if n.commitIndex.Id()<=n.recoverLogIndex&&n.commitIndex.Id()>n.stateMachine.lastApplied{
					//var lastApplied=n.stateMachine.lastApplied
					n.log.applyCommited()
					//Tracef("Node.Commit %s lastApplied %d==>%d",n.address,lastApplied,n.stateMachine.lastApplied)
				}
				break
			}
		}
	}
	return nil
}