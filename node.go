// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package raft implements the Raft distributed consensus protocol.
package raft

import (
	"fmt"
	"github.com/hslam/atomic"
	"github.com/hslam/timer"
	"runtime"
	"sync"
	"time"
)

// Node is a raft node.
type Node interface {
	Start()
	Running() bool
	Stop()
	Stoped() bool
	State() string
	Leader() string
	Ready() bool
	Lease() bool
	LeaseRead() bool
	IsLeader() bool
	Address() string
	SetCodec(codec Codec)
	SetContext(context interface{})
	Context() interface{}
	SetGzipSnapshot(gzip bool)
	SetSnapshotPolicy(snapshotPolicy SnapshotPolicy)
	SetSnapshot(snapshot Snapshot)
	ClearSyncType()
	AppendSyncType(seconds, changes int)
	SetSyncTypes(saves []*SyncType)
	RegisterCommand(command Command) error
	Do(command Command) (interface{}, error)
	ReadIndex() bool
	Peers() []string
	Join(info *NodeInfo) (success bool)
	Leave(Address string) (success bool, ok bool)
	LookupPeer(addr string) *NodeInfo
	LeaderChange(leaderChange func())
	MemberChange(memberChange func())
}

type node struct {
	mu        sync.RWMutex
	nodesMut  sync.RWMutex
	syncMut   sync.RWMutex
	waitGroup sync.WaitGroup
	onceStart sync.Once
	onceStop  sync.Once

	running bool
	done    chan struct{}
	stoped  bool

	host    string
	port    int
	address string
	leader  string

	//config
	heartbeatTick time.Duration
	storage       *storage
	raft          Raft
	cluster       Cluster
	rpcs          *rpcs
	log           *waLog
	readIndex     *readIndex
	stateMachine  *stateMachine

	peers           map[string]*peer
	detectTicker    *time.Ticker
	keepAliveTicker *time.Ticker
	printTicker     *time.Ticker
	checkLogTicker  *time.Ticker
	state           state
	changeStateChan chan int
	ticker          *time.Ticker
	updateTicker    *time.Ticker
	workTicker      *timer.Ticker
	deferTime       time.Time
	//persistent state on all servers
	currentTerm *atomic.Uint64
	votedFor    *atomic.String

	//volatile state on all servers
	commitIndex   *persistentUint64
	firstLogIndex uint64
	lastLogIndex  uint64
	lastLogTerm   uint64
	nextIndex     uint64

	//init
	recoverLogIndex        uint64
	lastPrintFirstLogIndex uint64
	lastPrintLastLogIndex  uint64
	lastPrintCommitIndex   uint64
	lastPrintLastApplied   uint64
	lastPrintNextIndex     uint64

	//candidate
	votes    *votes
	election *election

	//leader
	lease bool
	ready bool

	codec     Codec
	raftCodec Codec

	context  interface{}
	commands *commands
	pipeline *pipeline

	commiting int32

	nonVoting  bool
	majorities bool
	leave      bool

	leaderChange func()
	memberChange func()
}

// NewNode returns a new raft node.
func NewNode(host string, port int, dataDir string, context interface{}, join bool, nodes []*NodeInfo) (Node, error) {
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	address := fmt.Sprintf("%s:%d", host, port)
	n := &node{
		host:            host,
		port:            port,
		address:         address,
		storage:         newStorage(dataDir),
		peers:           make(map[string]*peer),
		printTicker:     time.NewTicker(defaultNodeTracePrintTick),
		detectTicker:    time.NewTicker(defaultDetectTick),
		keepAliveTicker: time.NewTicker(defaultKeepAliveTick),
		checkLogTicker:  time.NewTicker(defaultCheckLogTick),
		ticker:          time.NewTicker(defaultNodeTick),
		updateTicker:    time.NewTicker(defaultUpdateTick),
		done:            make(chan struct{}, 1),
		changeStateChan: make(chan int, 1),
		heartbeatTick:   defaultHeartbeatTick,
		codec:           new(GOGOPBCodec),
		raftCodec:       new(GOGOPBCodec),
		context:         context,
		commands:        &commands{types: make(map[int32]*sync.Pool)},
		nextIndex:       1,
		currentTerm:     atomic.NewUint64(0),
		votedFor:        atomic.NewString(""),
	}
	n.votes = newVotes(n)
	n.readIndex = newReadIndex(n)
	n.stateMachine = newStateMachine(n)
	n.setNodes(nodes)
	n.log = newLog(n)
	n.election = newElection(n, defaultElectionTimeout)
	n.raft = newRaft(n)
	n.cluster = newCluster(n)
	n.rpcs = newRPCs(n, fmt.Sprintf(":%d", port))
	n.commitIndex = newPersistentUint64(n, defaultCommitIndex, 0, time.Second)
	n.state = newFollowerState(n)
	n.pipeline = newPipeline(n)
	n.registerCommand(&NoOperationCommand{})
	n.registerCommand(&AddPeerCommand{})
	n.registerCommand(&RemovePeerCommand{})
	n.registerCommand(&ReconfigurationCommand{})
	if join {
		n.majorities = false
		go func() {
			nodeInfo := n.stateMachine.configuration.LookupPeer(n.address)
			if nodeInfo == nil {
				return
			}
			for {
				time.Sleep(time.Second)
				if n.running {
					if n.Join(nodeInfo) {
						break
					}
				}
			}
		}()
	} else {
		n.majorities = true
	}
	return n, nil
}

func (n *node) Start() {
	n.onceStart.Do(func() {
		n.recover()
		n.currentTerm.Store(n.lastLogTerm + 1)
		n.checkLog()
		go n.rpcs.ListenAndServe()
		go n.run()
	})
	n.election.Reset()
	n.running = true
}

func (n *node) run() {
	for {
		runtime.Gosched()
		select {
		case v, ok := <-n.votes.vote:
			if ok {
				n.votes.AddVote(v)
			}
		case <-n.ticker.C:
			if !n.running {
				return
			}
			select {
			case i := <-n.changeStateChan:
				if i == 1 {
					n.setState(n.state.NextState())
				} else if i == -1 {
					n.setState(n.state.StepDown())
				} else if i == 0 {
					n.state.Start()
				}
			default:
				n.state.FixedUpdate()
			}
		case <-n.updateTicker.C:
			if !n.running {
				return
			}
			if n.workTicker != nil && n.deferTime.Add(time.Duration(minLatency*10)).Before(time.Now()) {
				n.workTicker.Stop()
				n.workTicker = nil
			}
			if n.state.Update() {
				if n.workTicker == nil {
					n.deferTime = time.Now()
					n.workTicker = timer.TickFunc(defaultMaxDelay, func() {
						n.state.Update()
					})
				}
			}
		case <-n.printTicker.C:
			n.print()
		case <-n.keepAliveTicker.C:
			n.keepAliveNodes()
		case <-n.detectTicker.C:
			n.detectNodes()
		case <-n.checkLogTicker.C:
			n.checkLog()
		case <-n.done:
			goto endfor
		}
	}
endfor:
	n.ticker.Stop()
	n.updateTicker.Stop()
	n.printTicker.Stop()
	n.keepAliveTicker.Stop()
	n.detectTicker.Stop()
	n.checkLogTicker.Stop()
}

func (n *node) Running() bool {
	n.mu.RLock()
	running := n.running
	n.mu.RUnlock()
	return running
}

func (n *node) Stop() {
	n.onceStop.Do(func() {
		close(n.done)
		n.stoped = true
		n.running = false
		n.commitIndex.Stop()
		n.rpcs.Stop()
		n.stateMachine.Stop()
		n.pipeline.Stop()
		n.readIndex.Stop()
	})
}

func (n *node) Stoped() bool {
	n.mu.RLock()
	stoped := n.stoped
	n.mu.RUnlock()
	return stoped
}

func (n *node) voting() bool {
	return !n.nonVoting && n.majorities
}

func (n *node) setState(s state) {
	n.state = s
}

func (n *node) stepDown() {
	n.changeState(-1)
}

func (n *node) nextState() {
	n.changeState(1)
}

func (n *node) stay() {
	n.changeState(0)
}

func (n *node) changeState(i int) {
	if len(n.changeStateChan) == 1 {
		return
	}
	n.changeStateChan <- i
}

func (n *node) State() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.state == nil {
		return ""
	}
	return n.state.String()
}

func (n *node) Term() uint64 {
	n.mu.RLock()
	term := n.term()
	n.mu.RUnlock()
	return term
}

func (n *node) term() uint64 {
	if !n.running {
		return 0
	}
	return n.currentTerm.Load()
}

func (n *node) LeaderChange(leaderChange func()) {
	n.mu.RLock()
	n.leaderChange = leaderChange
	n.mu.RUnlock()
}

func (n *node) MemberChange(memberChange func()) {
	n.mu.RLock()
	n.memberChange = memberChange
	n.mu.RUnlock()
}

func (n *node) Leader() string {
	n.mu.RLock()
	leader := n.leader
	n.mu.RUnlock()
	return leader
}

func (n *node) Ready() bool {
	n.mu.RLock()
	ready := n.ready
	n.mu.RUnlock()
	return ready
}

func (n *node) Lease() bool {
	n.mu.RLock()
	lease := n.lease
	n.mu.RUnlock()
	return lease
}

func (n *node) LeaseRead() (ok bool) {
	commitIndex := n.commitIndex.ID()
	if atomic.LoadUint64(&n.stateMachine.lastApplied) < commitIndex {
		timer := time.NewTimer(defaultCommandTimeout)
		var done = make(chan struct{}, 1)
		go n.waitApply(commitIndex, done)
		select {
		case <-done:
			timer.Stop()
		case <-timer.C:
			ok = false
			return
		}
	}
	return n.Lease()
}

func (n *node) IsLeader() bool {
	n.mu.RLock()
	isLeader := n.isLeader()
	n.mu.RUnlock()
	return isLeader
}

func (n *node) isLeader() bool {
	if !n.running {
		return false
	}
	if n.leader == n.address && n.state.String() == leader && n.lease {
		return true
	}
	return false
}

func (n *node) Address() string {
	n.mu.RLock()
	address := n.address
	n.mu.RUnlock()
	return address
}

func (n *node) SetCodec(codec Codec) {
	n.mu.RLock()
	n.codec = codec
	n.mu.RUnlock()
}

func (n *node) SetContext(context interface{}) {
	n.mu.Lock()
	n.context = context
	n.mu.Unlock()
}

func (n *node) Context() interface{} {
	n.mu.RLock()
	context := n.context
	n.mu.RUnlock()
	return context
}

func (n *node) SetGzipSnapshot(gzip bool) {
	n.stateMachine.snapshotReadWriter.Gzip(gzip)
}

func (n *node) SetSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	n.stateMachine.SetSnapshotPolicy(snapshotPolicy)
}

func (n *node) SetSnapshot(snapshot Snapshot) {
	n.stateMachine.SetSnapshot(snapshot)
}

func (n *node) ClearSyncType() {
	n.stateMachine.ClearSyncType()
}

func (n *node) AppendSyncType(seconds, changes int) {
	n.stateMachine.AppendSyncType(seconds, changes)
}

func (n *node) SetSyncTypes(saves []*SyncType) {
	n.stateMachine.SetSyncTypes(saves)
}

func (n *node) RegisterCommand(command Command) error {
	if command == nil {
		return ErrCommandNil
	} else if command.Type() < 0 {
		return ErrCommandTypeMinus
	}
	n.mu.Lock()
	err := n.commands.register(command)
	n.mu.Unlock()
	return err
}

func (n *node) registerCommand(command Command) error {
	if command == nil {
		return ErrCommandNil
	}
	return n.commands.register(command)
}

func (n *node) Do(command Command) (interface{}, error) {
	if command.Type() < 0 {
		return nil, ErrCommandTypeMinus
	}
	return n.do(command, defaultCommandTimeout)
}

func (n *node) do(command Command, timeout time.Duration) (reply interface{}, err error) {
	i := n.put(command)
	if i.Error == nil {
		timer := time.NewTimer(timeout)
		runtime.Gosched()
		select {
		case <-i.Done:
			timer.Stop()
			reply = i.Reply
			err = i.Error
		case <-timer.C:
			err = ErrCommandTimeout
		}
	} else {
		reply = i.Reply
		err = i.Error
	}
	freeInvoker(i)
	return
}

func (n *node) put(command Command) *invoker {
	var i = newInvoker()
	i.Command = command
	if !n.running {
		i.Error = ErrNotRunning
		i.done()
		return i
	}
	if command == nil {
		i.Error = ErrCommandNil
		i.done()
		return i
	}
	if n.IsLeader() {
		if n.commands.exists(command) {
			n.pipeline.write(i)
			return i
		}
		i.Error = ErrCommandNotRegistered
	} else {
		i.Error = ErrNotLeader
	}
	i.done()
	return i
}

func (n *node) ReadIndex() bool {
	if !n.running {
		return false
	}
	if !n.isLeader() {
		return false
	}
	if !n.Ready() {
		return false
	}
	return n.readIndex.Read()
}

func (n *node) waitApply(commitIndex uint64, done chan struct{}) {
	for {
		if atomic.LoadUint64(&n.stateMachine.lastApplied) >= commitIndex {
			select {
			case done <- struct{}{}:
			default:
			}
			return
		}
		time.Sleep(n.readIndex.minLatency() / 10)
	}
}

func (n *node) Peers() []string {
	n.nodesMut.Lock()
	peers := make([]string, 0, len(n.peers))
	for _, v := range n.peers {
		peers = append(peers, v.address)
	}
	n.nodesMut.Unlock()
	return peers
}

func (n *node) membership() []string {
	n.nodesMut.Lock()
	ms := make([]string, 0, len(n.peers)+1)
	if !n.leave {
		ms = append(ms, fmt.Sprintf("%s;%t", n.address, n.nonVoting))
	}
	for _, v := range n.peers {
		ms = append(ms, fmt.Sprintf("%s;%t", v.address, v.nonVoting))
	}
	n.nodesMut.Unlock()
	return ms
}

func (n *node) Join(info *NodeInfo) (success bool) {
	leader := n.Leader()
	if leader != "" {
		success, ok := n.cluster.CallAddPeer(leader, info)
		if success && ok {
			return true
		}
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := n.cluster.CallQueryLeader(peers[i])
		if leaderID != "" && ok {
			success, ok := n.cluster.CallAddPeer(leaderID, info)
			if success && ok {
				return true
			}
		}
	}
	return false
}

func (n *node) Leave(Address string) (success bool, ok bool) {
	leader := n.Leader()
	if leader != "" {
		return n.cluster.CallRemovePeer(leader, Address)
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := n.cluster.CallQueryLeader(peers[i])
		if leaderID != "" && ok {
			n.cluster.CallRemovePeer(leaderID, Address)
		}
	}
	return
}

func (n *node) setNodes(nodes []*NodeInfo) {
	n.stateMachine.configuration.SetNodes(nodes)
	n.stateMachine.configuration.load()
	for _, v := range n.peers {
		v.majorities = true
	}
	n.resetVotes()
}

func (n *node) addNode(info *NodeInfo) {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	if _, ok := n.peers[info.Address]; ok {
		n.peers[info.Address].nonVoting = info.NonVoting
		return
	}
	if n.address != info.Address {
		peer := newPeer(n, info.Address)
		peer.nonVoting = info.NonVoting
		n.peers[info.Address] = peer
	} else {
		n.nonVoting = info.NonVoting
	}
	n.resetVotes()
	return
}

func (n *node) resetVotes() {
	if n.votingsCount() == 0 {
		n.votes.Reset(1)
	}
	n.votes.Reset(n.votingsCount())
	return
}

func (n *node) consideredForMajorities() {
	n.nodesMut.Lock()
	if n.stateMachine.configuration.LookupPeer(n.address) != nil {
		n.majorities = true
	} else {
		n.majorities = false
	}
	for _, v := range n.peers {
		v.majorities = true
	}
	n.nodesMut.Unlock()
}

func (n *node) deleteNotPeers(peers []string) {
	if len(peers) == 0 {
		n.clearPeers()
		return
	}
	m := make(map[string]bool)
	for _, v := range peers {
		m[v] = true
	}
	n.nodesMut.Lock()
	for _, v := range n.peers {
		if _, ok := m[v.address]; !ok {
			delete(n.peers, v.address)
		}
	}
	n.nodesMut.Unlock()
}

func (n *node) clearPeers() {
	n.nodesMut.Lock()
	n.peers = make(map[string]*peer)
	n.nodesMut.Unlock()
}

func (n *node) LookupPeer(addr string) *NodeInfo {
	n.nodesMut.Lock()
	nodeInfo := n.stateMachine.configuration.LookupPeer(addr)
	n.nodesMut.Unlock()
	return nodeInfo
}

func (n *node) NodesCount() int {
	n.nodesMut.RLock()
	count := len(n.peers) + 1
	n.nodesMut.RUnlock()
	return count
}

func (n *node) Quorum() int {
	n.nodesMut.RLock()
	quorum := n.quorum()
	n.nodesMut.RUnlock()
	return quorum
}

func (n *node) quorum() int {
	return n.votingsCount()/2 + 1
}

func (n *node) votingsCount() int {
	cnt := 0
	for _, v := range n.peers {
		if v.voting() {
			cnt++
		}
	}
	if n.voting() {
		cnt++
	}
	return cnt
}

func (n *node) AliveCount() int {
	n.nodesMut.RLock()
	aliveCount := n.aliveCount()
	n.nodesMut.RUnlock()
	return aliveCount
}

func (n *node) aliveCount() int {
	cnt := 1
	for _, v := range n.peers {
		if v.alive == true && v.voting() {
			cnt++
		}
	}
	return cnt
}

func (n *node) requestVotes() error {
	n.nodesMut.RLock()
	n.votes.Clear()
	n.votes.vote <- newVote(n.address, n.currentTerm.Load(), 1)
	for _, v := range n.peers {
		if v.alive == true && v.voting() {
			go v.requestVote()
		}
	}
	n.nodesMut.RUnlock()
	return nil
}

func (n *node) detectNodes() error {
	n.nodesMut.RLock()
	for _, v := range n.peers {
		if v.alive == false {
			go v.ping()
		}
	}
	n.nodesMut.RUnlock()
	return nil
}

func (n *node) keepAliveNodes() error {
	n.nodesMut.RLock()
	for _, v := range n.peers {
		go v.ping()
	}
	n.nodesMut.RUnlock()
	return nil
}

func (n *node) heartbeats() error {
	n.nodesMut.RLock()
	for _, v := range n.peers {
		if v.alive {
			go v.heartbeat()
		}
	}
	n.nodesMut.RUnlock()
	return nil
}

func (n *node) checkLeader() bool {
	n.nodesMut.RLock()
	peers := n.peers
	quorum := uint32(n.quorum())
	n.nodesMut.RUnlock()
	done := make(chan struct{}, 1)
	count := uint32(1)
	send := uint32(1)
	for i := range peers {
		v := peers[i]
		if !v.voting() {
			continue
		}
		if v.alive {
			atomic.AddUint32(&send, 1)
			go func() {
				ok := v.heartbeat()
				//logger.Tracef("node.checkLeader %s %v", n.address, ok)
				if ok {
					new := atomic.AddUint32(&count, 1)
					if new >= quorum {
						select {
						case done <- struct{}{}:
						default:
						}
					}
				}
			}()
		}
	}
	if send < quorum {
		return false
	}
	//logger.Tracef("node.checkLeader %s quorum-%v", n.address, quorum)
	timer := time.NewTimer(defaultCommandTimeout)
	runtime.Gosched()
	select {
	case <-done:
		timer.Stop()
		//logger.Tracef("node.checkLeader %s count-%v quorum-%v", n.address, count, quorum)
		return true
	case <-timer.C:
		//logger.Tracef("node.checkLeader %s timeout count-%v quorum-%v", n.address, count, quorum)
	}
	return false
}

func (n *node) install() bool {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if atomic.LoadInt32(&v.install) > 0 {
			return false
		}
	}
	return true
}

func (n *node) check() error {
	n.nodesMut.RLock()
	for _, v := range n.peers {
		if v.alive == true {
			go v.check()
		}
	}
	if len(n.peers) == 0 {
		go n.commit()
	}
	n.nodesMut.RUnlock()
	return nil
}

func (n *node) minNextIndex() uint64 {
	n.nodesMut.RLock()
	var min uint64
	for _, v := range n.peers {
		if v.alive == true && v.nextIndex > 0 {
			if min == 0 {
				min = v.nextIndex
			} else {
				min = minUint64(min, v.nextIndex)
			}
		}
	}
	n.nodesMut.RUnlock()
	return min
}

func (n *node) reset() {
	n.recoverLogIndex = 0
	n.lastPrintNextIndex = 1
	n.lastPrintLastApplied = 0
	n.lastPrintLastLogIndex = 0
	n.lastPrintCommitIndex = 0
	n.nextIndex = 1
	n.lastLogIndex = 0
	n.lastLogTerm = 0
	n.commitIndex.Set(0)
	n.stateMachine.lastApplied = 0
	n.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(0)
	n.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(0)
	n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
}

func (n *node) load() {
	n.commitIndex.load()
	n.stateMachine.load()
	n.log.load()
}

func (n *node) recover() error {
	logger.Tracef("node.recover %s start", n.address)
	n.load()
	ticker := time.NewTicker(time.Second)
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				n.print()
			case <-done:
				ticker.Stop()
				goto endfor
			}
		}
	endfor:
	}()
	n.stateMachine.recover()
	n.print()
	n.log.applyCommited()
	n.print()
	close(done)
	logger.Tracef("node.recover %s finish", n.address)
	return nil
}

func (n *node) checkLog() error {
	if n.storage.IsEmpty(defaultLastIncludedIndex) {
		n.stateMachine.snapshotReadWriter.lastIncludedIndex.save()
	}
	if n.storage.IsEmpty(defaultLastIncludedTerm) {
		n.stateMachine.snapshotReadWriter.lastIncludedTerm.save()
	}
	if n.storage.IsEmpty(defaultLastTarIndex) {
		n.stateMachine.snapshotReadWriter.lastTarIndex.save()
	}
	if n.storage.IsEmpty(defaultConfig) {
		n.stateMachine.configuration.save()
	}
	if n.storage.IsEmpty(defaultSnapshot) && n.stateMachine.snapshot != nil {
		n.stateMachine.SaveSnapshot()
	} else if n.stateMachine.snapshot == nil {
		if !n.storage.Exists(defaultSnapshot) {
			n.storage.Truncate(defaultSnapshot, 1)
		}
	}
	if n.isLeader() && n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) && !n.storage.IsEmpty(defaultSnapshot) && !n.storage.IsEmpty(defaultLastIncludedIndex) && !n.storage.IsEmpty(defaultLastIncludedTerm) {
		n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
		n.stateMachine.snapshotReadWriter.Tar()
	}
	return nil
}

func (n *node) printPeers() {
	if !n.isLeader() {
		return
	}
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.nextIndex > v.lastPrintNextIndex {
			logger.Tracef("node.printPeers %s nextIndex %d==>%d", v.address, v.lastPrintNextIndex, v.nextIndex)
			v.lastPrintNextIndex = v.nextIndex
		}
	}
}

func (n *node) print() {
	if n.firstLogIndex > n.lastPrintFirstLogIndex {
		logger.Tracef("node.print %s firstLogIndex %d==>%d", n.address, n.lastPrintFirstLogIndex, n.firstLogIndex)
		n.lastPrintFirstLogIndex = n.firstLogIndex
	}
	//if n.lastLogIndex > n.lastPrintLastLogIndex {
	//	logger.Tracef("node.print %s lastLogIndex %d==>%d", n.address, n.lastPrintLastLogIndex, n.lastLogIndex)
	//	n.lastPrintLastLogIndex = n.lastLogIndex
	//}
	//if n.commitIndex > n.lastPrintCommitIndex {
	//	logger.Tracef("node.print %s commitIndex %d==>%d", n.address, n.lastPrintCommitIndex, n.commitIndex)
	//	n.lastPrintCommitIndex = n.commitIndex
	//}
	if n.stateMachine.lastApplied > n.lastPrintLastApplied {
		logger.Tracef("node.print %s lastApplied %d==>%d", n.address, n.lastPrintLastApplied, n.stateMachine.lastApplied)
		n.lastPrintLastApplied = n.stateMachine.lastApplied
	}
	//if n.nextIndex > n.lastPrintNextIndex {
	//	logger.Tracef("node.print %s nextIndex %d==>%d", n.address, n.lastPrintNextIndex, n.nextIndex)
	//	n.lastPrintNextIndex = n.nextIndex
	//}
	n.printPeers()
}

func (n *node) commit() bool {
	if !atomic.CompareAndSwapInt32(&n.commiting, 0, 1) {
		return true
	}
	defer atomic.StoreInt32(&n.commiting, 0)
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	if n.votingsCount() == 1 {
		index := n.lastLogIndex
		if index > n.commitIndex.ID() {
			n.commitIndex.Set(index)
			go n.pipeline.apply()
			return true
		}
		return false
	}
	var lastLogIndexs = make([]uint64, 1, n.votingsCount())
	lastLogIndexs[0] = n.lastLogIndex
	for _, v := range n.peers {
		if !v.voting() {
			continue
		}
		if v.nextIndex > 0 {
			lastLogIndexs = append(lastLogIndexs, v.nextIndex-1)
		} else {
			lastLogIndexs = append(lastLogIndexs, 0)
		}
	}
	quickSort(lastLogIndexs, -999, -999)
	index := lastLogIndexs[len(lastLogIndexs)/2]
	if index > n.commitIndex.ID() {
		//logger.Tracef("node.commit %s sort after %v %d", n.address, lastLogIndexs, index)
		n.commitIndex.Set(index)
		go n.pipeline.apply()
		return true
	}
	return false
}
