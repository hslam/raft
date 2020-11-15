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
	proxy         Proxy
	rpcs          *rpcs
	server        *server
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
	commitIndex   uint64
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

	raftCodec Codec
	codec     Codec

	context      interface{}
	commands     *commands
	pipeline     *pipeline
	pipelineChan chan bool

	commiting int32

	join       bool
	nonVoting  bool
	majorities bool
	leave      bool
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
		rpcs:            newRPCs(),
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
		raftCodec:       new(GOGOPBCodec),
		codec:           new(JSONCodec),
		context:         context,
		commands:        &commands{types: make(map[int32]*sync.Pool)},
		nextIndex:       1,
		join:            join,
	}
	n.storage = newStorage(dataDir)
	n.votes = newVotes(n)
	n.readIndex = newReadIndex(n)
	n.stateMachine = newStateMachine(n)
	n.setNodes(nodes)
	n.log = newLog(n)
	n.election = newElection(n, defaultElectionTimeout)
	n.raft = newRaft(n)
	n.proxy = newProxy(n)
	n.server = newServer(n, fmt.Sprintf(":%d", port))
	//n.currentTerm = newPersistentUint64(n, defaultTerm, 0, 0)
	n.currentTerm = atomic.NewUint64(0)
	n.commitIndex = 0
	//n.votedFor = newPersistentString(n, defaultVoteFor)
	n.votedFor = atomic.NewString("")
	n.state = newFollowerState(n)
	n.pipeline = newPipeline(n)
	n.pipelineChan = make(chan bool, defaultMaxConcurrency)
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
		n.checkLog()
		n.server.listenAndServe()
		go n.run()
	})
	n.election.Reset()
	n.running = true
}

func (n *node) run() {
	go func(updateTicker *time.Ticker, done chan struct{}) {
		for {
			select {
			case <-done:
				goto endfor
			case <-n.updateTicker.C:
				func() {
					defer func() {
						if err := recover(); err != nil {
						}
					}()
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
								defer func() {
									if err := recover(); err != nil {
									}
								}()
								n.state.Update()
							})
						}
					}
				}()
			}
		}
	endfor:
	}(n.updateTicker, n.done)
	for {
		select {
		case <-n.ticker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
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
			}()
		case <-n.printTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				n.print()
			}()
		case <-n.keepAliveTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				n.keepAliveNodes()
			}()
		case <-n.detectTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				n.detectNodes()
			}()
		case <-n.checkLogTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				n.checkLog()
			}()
		case v := <-n.votes.vote:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				n.votes.AddVote(v)
			}()
		case <-n.done:
			goto endfor
		}
	}
endfor:
	n.printTicker.Stop()
	n.printTicker = nil
	n.keepAliveTicker.Stop()
	n.keepAliveTicker = nil
	n.detectTicker.Stop()
	n.detectTicker = nil
	n.ticker.Stop()
	n.ticker = nil
	n.updateTicker.Stop()
	n.updateTicker = nil
}

func (n *node) Running() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}

func (n *node) Stop() {
	n.onceStop.Do(func() {
		close(n.done)
		n.stoped = true
		n.running = false
	})
}

func (n *node) Stoped() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stoped
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
	defer n.mu.RUnlock()
	return n.term()
}

func (n *node) term() uint64 {
	if !n.running {
		return 0
	}
	return n.currentTerm.Load()
}

func (n *node) Leader() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leader
}

func (n *node) Ready() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.ready
}

func (n *node) Lease() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lease
}

func (n *node) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isLeader()
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
	defer n.mu.RUnlock()
	return n.address
}

func (n *node) SetCodec(codec Codec) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.codec = codec
}

func (n *node) SetContext(context interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.context = context
}

func (n *node) Context() interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.context
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
	defer n.mu.Unlock()
	return n.commands.register(command)
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
	runtime.Gosched()
	if i.Error == nil {
		timer := time.NewTimer(timeout)
		select {
		case <-i.Done:
			timer.Stop()
			reply = i.Reply
			err = i.Error
		case <-timer.C:
			err = ErrCommandTimeout
		}
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

func (n *node) Peers() []string {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	peers := make([]string, 0, len(n.peers))
	for _, v := range n.peers {
		peers = append(peers, v.address)
	}
	return peers
}

func (n *node) membership() []string {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	ms := make([]string, 0, len(n.peers)+1)
	if !n.leave {
		ms = append(ms, fmt.Sprintf("%s;%t", n.address, n.nonVoting))
	}
	for _, v := range n.peers {
		ms = append(ms, fmt.Sprintf("%s;%t", v.address, v.nonVoting))
	}
	return ms
}

func (n *node) Join(info *NodeInfo) (success bool) {
	leader := n.Leader()
	if leader != "" {
		success, ok := n.proxy.AddPeer(leader, info)
		if success && ok {
			return true
		}
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := n.proxy.QueryLeader(peers[i])
		if leaderID != "" && ok {
			success, ok := n.proxy.AddPeer(leaderID, info)
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
		return n.proxy.RemovePeer(leader, Address)
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := n.proxy.QueryLeader(peers[i])
		if leaderID != "" && ok {
			n.proxy.RemovePeer(leaderID, Address)
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
	defer n.nodesMut.Unlock()
	if n.stateMachine.configuration.LookupPeer(n.address) != nil {
		n.majorities = true
	} else {
		n.majorities = false
	}
	for _, v := range n.peers {
		v.majorities = true
	}
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
	defer n.nodesMut.Unlock()
	for _, v := range n.peers {
		if _, ok := m[v.address]; !ok {
			delete(n.peers, v.address)
		}
	}
}

func (n *node) clearPeers() {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	n.peers = make(map[string]*peer)
}

func (n *node) LookupPeer(addr string) *NodeInfo {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	return n.stateMachine.configuration.LookupPeer(addr)
}

func (n *node) NodesCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return len(n.peers) + 1
}

func (n *node) Quorum() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return n.quorum()
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
	defer n.nodesMut.RUnlock()
	return n.aliveCount()
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
	defer n.nodesMut.RUnlock()
	n.votes.Clear()
	n.votes.vote <- newVote(n.address, n.currentTerm.Load(), 1)
	for _, v := range n.peers {
		if v.alive == true && v.voting() {
			go v.requestVote()
		}
	}
	return nil
}

func (n *node) detectNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive == false {
			go v.ping()
		}
	}
	return nil
}

func (n *node) keepAliveNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		go v.ping()
	}
	return nil
}

func (n *node) heartbeats() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive {
			go v.heartbeat()
		}
	}
	return nil
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
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive == true {
			go v.check()
		}
	}
	if len(n.peers) == 0 {
		go n.commit()
	}
	return nil
}

func (n *node) minNextIndex() uint64 {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
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
	n.commitIndex = 0
	n.stateMachine.lastApplied = 0
	n.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(0)
	n.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(0)
	n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
}

func (n *node) load() {
	n.stateMachine.load()
	n.log.load()
}

func (n *node) recover() error {
	logger.Tracef("node.recover %s start", n.address)
	//if n.storage.IsEmpty(DefaultIndex) {
	//	if !n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) {
	//		n.stateMachine.snapshotReadWriter.untar()
	//	}
	//} else if n.storage.IsEmpty(DefaultLog) {
	//	if !n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) {
	//		n.stateMachine.snapshotReadWriter.untar()
	//	}
	//}
	n.load()
	recoverApplyTicker := time.NewTicker(time.Second)
	recoverApplyStop := make(chan bool, 1)
	go func() {
		for {
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
	n.print()
	n.log.applyCommited()
	n.print()
	recoverApplyStop <- true
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
	//if n.storage.IsEmpty(defaultTerm) {
	//	n.currentTerm.save()
	//}
	if n.storage.IsEmpty(defaultConfig) {
		n.stateMachine.configuration.save()
	}
	//if n.storage.IsEmpty(defaultVoteFor) {
	//	n.votedFor.save()
	//}
	//if n.storage.IsEmpty(DefaultSnapshot) && n.stateMachine.snapshot != nil && !n.storage.IsEmpty(DefaultIndex) && !n.storage.IsEmpty(DefaultLog) {
	//	n.stateMachine.SaveSnapshot()
	//} else if n.stateMachine.snapshot == nil {
	//	if !n.storage.Exists(DefaultSnapshot) {
	//		n.storage.Truncate(DefaultSnapshot, 1)
	//	}
	//}
	//if n.isLeader() && n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) && !n.storage.IsEmpty(DefaultIndex) && !n.storage.IsEmpty(DefaultLog) && !n.storage.IsEmpty(DefaultSnapshot) && !n.storage.IsEmpty(DefaultLastIncludedIndex) && !n.storage.IsEmpty(DefaultLastIncludedTerm) {
	//	n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
	//	n.stateMachine.snapshotReadWriter.Tar()
	//}
	//if n.storage.IsEmpty(DefaultIndex) {
	//	if !n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) {
	//		n.stateMachine.snapshotReadWriter.untar()
	//		n.load()
	//	} else {
	//		n.nextIndex = 1
	//	}
	//} else if n.storage.IsEmpty(DefaultLog) {
	//	if !n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) {
	//		n.stateMachine.snapshotReadWriter.untar()
	//		n.load()
	//	} else {
	//		n.nextIndex = 1
	//	}
	//}
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
		if index > n.commitIndex {
			//n.storage.Sync(n.log.name)
			//n.storage.Sync(n.log.indexs.name)
			n.commitIndex = index
			//n.pipeline.commitIndex <- index
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
	if index > n.commitIndex {
		//logger.Tracef("node.commit %s sort after %v %d", n.address, lastLogIndexs, index)
		//n.storage.Sync(n.log.name)
		//n.storage.Sync(n.log.indexs.name)
		n.commitIndex = index
		//n.pipeline.commitIndex <- index
		go n.pipeline.apply()
		return true
	}
	return false
}
