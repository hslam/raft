// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package raft implements the Raft distributed consensus protocol.
package raft

import (
	"fmt"
	"github.com/hslam/timer"
	"sync"
	"time"
)

func init() {
}

type Node struct {
	mu        sync.RWMutex
	nodesMut  sync.RWMutex
	syncMut   sync.RWMutex
	waitGroup sync.WaitGroup
	onceStart sync.Once
	onceStop  sync.Once

	running bool
	stop    chan bool
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
	server        *Server
	log           *Log
	readIndex     *readIndex
	stateMachine  *stateMachine

	peers           map[string]*Peer
	detectTicker    *time.Ticker
	keepAliveTicker *time.Ticker
	printTicker     *time.Ticker
	checkLogTicker  *time.Ticker
	state           State
	changeStateChan chan int
	ticker          *time.Ticker
	updateTicker    *time.Ticker
	workTicker      *timer.Ticker
	deferTime       time.Time
	//persistent state on all servers
	currentTerm *persistentUint64
	votedFor    *persistentString

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

	commitWork bool
	join       bool
	nonVoting  bool
	majorities bool
	leave      bool
}

func NewNode(host string, port int, data_dir string, context interface{}, join bool, nodes []*NodeInfo) (*Node, error) {
	if data_dir == "" {
		data_dir = DefaultDataDir
	}
	address := fmt.Sprintf("%s:%d", host, port)
	n := &Node{
		host:            host,
		port:            port,
		address:         address,
		rpcs:            newRPCs(),
		peers:           make(map[string]*Peer),
		printTicker:     time.NewTicker(DefaultNodeTracePrintTick),
		detectTicker:    time.NewTicker(DefaultDetectTick),
		keepAliveTicker: time.NewTicker(DefaultKeepAliveTick),
		checkLogTicker:  time.NewTicker(DefaultCheckLogTick),
		ticker:          time.NewTicker(DefaultNodeTick),
		updateTicker:    time.NewTicker(DefaultUpdateTick),
		stop:            make(chan bool, 1),
		changeStateChan: make(chan int, 1),
		heartbeatTick:   DefaultHeartbeatTick,
		raftCodec:       new(GOGOPBCodec),
		codec:           new(JSONCodec),
		context:         context,
		commands:        &commands{types: make(map[int32]*sync.Pool)},
		nextIndex:       1,
		join:            join,
		commitWork:      true,
	}
	n.storage = newStorage(n, data_dir)
	n.votes = newVotes(n)
	n.readIndex = newReadIndex(n)
	n.stateMachine = newStateMachine(n)
	n.setNodes(nodes)
	n.log = newLog(n)
	n.election = newElection(n, DefaultElectionTimeout)
	n.raft = newRaft(n)
	n.proxy = newProxy(n)
	n.server = newServer(n, fmt.Sprintf(":%d", port))
	n.currentTerm = newPersistentUint64(n, DefaultTerm, 0)
	n.commitIndex = 0
	n.votedFor = newPersistentString(n, DefaultVoteFor)
	n.state = newFollowerState(n)
	n.pipeline = newPipeline(n)
	n.pipelineChan = make(chan bool, DefaultMaxConcurrency)
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
func (n *Node) Start() {
	n.onceStart.Do(func() {
		n.recover()
		n.checkLog()
		n.server.listenAndServe()
		go n.run()
	})
	n.election.Reset()
	n.running = true
}
func (n *Node) run() {
	updateStop := make(chan bool, 1)
	go func(updateTicker *time.Ticker, updateStop chan bool) {
		for {
			select {
			case <-updateStop:
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
					if n.workTicker != nil && n.deferTime.Add(DefaultCommandTimeout).Before(time.Now()) {
						n.workTicker.Stop()
						n.workTicker = nil
					}
					if n.state.Update() || n.pipeline.Update() || n.readIndex.Update() {
						if n.workTicker == nil {
							n.deferTime = time.Now()
							n.workTicker = timer.TickFunc(DefaultMaxDelay, func() {
								defer func() {
									if err := recover(); err != nil {
									}
								}()
								n.state.Update()
								n.pipeline.Update()
								n.readIndex.Update()
							})
						}
					}
				}()
			}
		}
	endfor:
	}(n.updateTicker, updateStop)
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
		case <-n.stop:
			updateStop <- true
			goto endfor
		}
	}
endfor:
	close(n.stop)
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

func (n *Node) Running() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.running
}
func (n *Node) Stop() {
	n.onceStop.Do(func() {
		n.stop <- true
		n.stoped = true
	})
}
func (n *Node) Stoped() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stoped
}
func (n *Node) voting() bool {
	return !n.nonVoting && n.majorities
}

func (n *Node) setState(state State) {
	n.state = state
}

func (n *Node) GetState() State {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state
}

func (n *Node) stepDown() {
	n.changeState(-1)
}
func (n *Node) nextState() {
	n.changeState(1)
}

func (n *Node) stay() {
	n.changeState(0)
}

func (n *Node) changeState(i int) {
	if len(n.changeStateChan) == 1 {
		return
	}
	n.changeStateChan <- i
}

func (n *Node) State() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.state == nil {
		return ""
	}
	return n.state.String()
}

func (n *Node) Term() uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.term()
}
func (n *Node) term() uint64 {
	if !n.running {
		return 0
	}
	return n.currentTerm.Id()
}
func (n *Node) Leader() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leader
}
func (n *Node) Ready() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.ready
}
func (n *Node) Lease() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lease
}
func (n *Node) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isLeader()
}
func (n *Node) isLeader() bool {
	if !n.running {
		return false
	}
	if n.leader == n.address && n.state.String() == Leader && n.lease {
		return true
	} else {
		return false
	}
}
func (n *Node) Address() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.address
}
func (n *Node) SetCodec(codec Codec) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.codec = codec
}
func (n *Node) SetContext(context interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.context = context
}

func (n *Node) Context() interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.context
}

func (n *Node) GzipSnapshot() {
	n.stateMachine.snapshotReadWriter.Gzip(true)
}

func (n *Node) SetSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	n.stateMachine.SetSnapshotPolicy(snapshotPolicy)
}

func (n *Node) SetSnapshot(snapshot Snapshot) {
	n.stateMachine.SetSnapshot(snapshot)
}

func (n *Node) ClearSyncType() {
	n.stateMachine.ClearSyncType()
}

func (n *Node) AppendSyncType(seconds, changes int) {
	n.stateMachine.AppendSyncType(seconds, changes)
}

func (n *Node) SetSyncTypes(saves []*SyncType) {
	n.stateMachine.SetSyncTypes(saves)
}

func (n *Node) RegisterCommand(command Command) error {
	if command == nil {
		return ErrCommandNil
	} else if command.Type() < 0 {
		return ErrCommandTypeMinus
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.commands.register(command)
}
func (n *Node) registerCommand(command Command) error {
	if command == nil {
		return ErrCommandNil
	}
	return n.commands.register(command)
}
func (n *Node) Do(command Command) (interface{}, error) {
	if command.Type() < 0 {
		return nil, ErrCommandTypeMinus
	}
	return n.do(command, DefaultCommandTimeout)
}
func (n *Node) do(command Command, timeout time.Duration) (reply interface{}, err error) {
	if !n.running {
		return nil, ErrNotRunning
	}
	if command == nil {
		return nil, ErrCommandNil
	}
	if n.IsLeader() {
		if n.commands.exists(command) {
			var invoker = invokerPool.Get().(*Invoker)
			var done = donePool.Get().(chan *Invoker)
			invoker.Command = command
			invoker.Reply = reply
			invoker.Error = err
			invoker.Done = done
			n.pipeline.write(invoker)
			select {
			case <-done:
				for len(done) > 0 {
					<-done
				}
				donePool.Put(done)
				reply = invoker.Reply
				err = invoker.Error
				*invoker = Invoker{}
				invokerPool.Put(invoker)
			case <-time.After(timeout):
				err = ErrCommandTimeout
			}
		} else {
			err = ErrCommandNotRegistered
		}
	} else {
		err = ErrNotLeader
	}
	return reply, err
}
func (n *Node) ReadIndex() bool {
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
func (n *Node) fast() bool {
	return n.isLeader() && n.votingsCount() > 1
}
func (n *Node) Peers() []string {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	peers := make([]string, 0, len(n.peers))
	for _, v := range n.peers {
		peers = append(peers, v.address)
	}
	return peers
}
func (n *Node) membership() []string {
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
func (n *Node) Join(info *NodeInfo) (success bool) {
	leader := n.Leader()
	if leader != "" {
		success, ok := n.proxy.AddPeer(leader, info)
		if success && ok {
			return true
		}
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderId, ok := n.proxy.QueryLeader(peers[i])
		if leaderId != "" && ok {
			success, ok := n.proxy.AddPeer(leaderId, info)
			if success && ok {
				return true
			}
		}
	}
	return false
}
func (n *Node) Leave(Address string) (success bool, ok bool) {
	leader := n.Leader()
	if leader != "" {
		return n.proxy.RemovePeer(leader, Address)
	}
	peers := n.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderId, ok := n.proxy.QueryLeader(peers[i])
		if leaderId != "" && ok {
			n.proxy.RemovePeer(leaderId, Address)
		}
	}
	return
}
func (n *Node) setNodes(nodes []*NodeInfo) {
	n.stateMachine.configuration.SetNodes(nodes)
	n.stateMachine.configuration.load()
	for _, v := range n.peers {
		v.majorities = true
	}
	n.resetVotes()
}
func (n *Node) addNode(info *NodeInfo) {
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
func (n *Node) resetVotes() {
	if n.votingsCount() == 0 {
		n.votes.Reset(1)
	}
	n.votes.Reset(n.votingsCount())
	return
}
func (n *Node) consideredForMajorities() {
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
func (n *Node) deleteNotPeers(peers []string) {
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
func (n *Node) clearPeers() {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	n.peers = make(map[string]*Peer)
}
func (n *Node) LookupPeer(addr string) *NodeInfo {
	n.nodesMut.Lock()
	defer n.nodesMut.Unlock()
	return n.stateMachine.configuration.LookupPeer(addr)
}
func (n *Node) NodesCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return len(n.peers) + 1
}
func (n *Node) Quorum() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return n.quorum()
}
func (n *Node) quorum() int {
	return n.votingsCount()/2 + 1
}
func (n *Node) votingsCount() int {
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
func (n *Node) AliveCount() int {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	return n.aliveCount()
}
func (n *Node) aliveCount() int {
	cnt := 1
	for _, v := range n.peers {
		if v.alive == true && v.voting() {
			cnt++
		}
	}
	return cnt
}
func (n *Node) requestVotes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	n.votes.Clear()
	n.votes.vote <- newVote(n.address, n.currentTerm.Id(), 1)
	for _, v := range n.peers {
		if v.alive == true && v.voting() {
			go v.requestVote()
		}
	}
	return nil
}
func (n *Node) detectNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive == false {
			go v.ping()
		}
	}
	return nil
}
func (n *Node) keepAliveNodes() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		go v.ping()
	}
	return nil
}
func (n *Node) heartbeats() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive {
			go v.heartbeat()
		}
	}
	return nil
}
func (n *Node) install() bool {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if !v.install {
			return false
		}
	}
	return true
}
func (n *Node) check() error {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.alive == true {
			v.check()
		}
	}
	return nil
}
func (n *Node) minNextIndex() uint64 {
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
func (n *Node) reset() {
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

func (n *Node) load() {
	n.stateMachine.load()
	n.log.load()
}

func (n *Node) recover() error {
	Tracef("Node.recover %s start", n.address)
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
	Tracef("Node.recover %s finish", n.address)
	return nil
}
func (n *Node) checkLog() error {
	if n.storage.IsEmpty(DefaultLastIncludedIndex) {
		n.stateMachine.snapshotReadWriter.lastIncludedIndex.save()
	}
	if n.storage.IsEmpty(DefaultLastIncludedTerm) {
		n.stateMachine.snapshotReadWriter.lastIncludedTerm.save()
	}
	if n.storage.IsEmpty(DefaultLastTarIndex) {
		n.stateMachine.snapshotReadWriter.lastTarIndex.save()
	}
	if n.storage.IsEmpty(DefaultTerm) {
		n.currentTerm.save()
	}
	if n.storage.IsEmpty(DefaultConfig) {
		n.stateMachine.configuration.save()
	}
	if n.storage.IsEmpty(DefaultVoteFor) {
		n.votedFor.save()
	}
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
	if n.storage.IsEmpty(DefaultSnapshot) && n.stateMachine.snapshot != nil {
		n.stateMachine.SaveSnapshot()
	} else if n.stateMachine.snapshot == nil {
		if !n.storage.Exists(DefaultSnapshot) {
			n.storage.Truncate(DefaultSnapshot, 1)
		}
	}
	if n.isLeader() && n.storage.IsEmpty(n.stateMachine.snapshotReadWriter.FileName()) && !n.storage.IsEmpty(DefaultSnapshot) && !n.storage.IsEmpty(DefaultLastIncludedIndex) && !n.storage.IsEmpty(DefaultLastIncludedTerm) {
		n.stateMachine.snapshotReadWriter.lastTarIndex.Set(0)
		n.stateMachine.snapshotReadWriter.Tar()
	}
	return nil
}

func (n *Node) printPeers() {
	if !n.isLeader() {
		return
	}
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	for _, v := range n.peers {
		if v.nextIndex > v.lastPrintNextIndex {
			Tracef("Node.printPeers %s nextIndex %d==>%d", v.address, v.lastPrintNextIndex, v.nextIndex)
			v.lastPrintNextIndex = v.nextIndex
		}
	}
}

func (n *Node) print() {
	if n.firstLogIndex > n.lastPrintFirstLogIndex {
		Infof("Node.print %s firstLogIndex %d==>%d", n.address, n.lastPrintFirstLogIndex, n.firstLogIndex)
		n.lastPrintFirstLogIndex = n.firstLogIndex
	}
	if n.lastLogIndex > n.lastPrintLastLogIndex {
		Infof("Node.print %s lastLogIndex %d==>%d", n.address, n.lastPrintLastLogIndex, n.lastLogIndex)
		n.lastPrintLastLogIndex = n.lastLogIndex
	}
	if n.commitIndex > n.lastPrintCommitIndex {
		Infof("Node.print %s commitIndex %d==>%d", n.address, n.lastPrintCommitIndex, n.commitIndex)
		n.lastPrintCommitIndex = n.commitIndex
	}
	if n.stateMachine.lastApplied > n.lastPrintLastApplied {
		Infof("Node.print %s lastApplied %d==>%d", n.address, n.lastPrintLastApplied, n.stateMachine.lastApplied)
		n.lastPrintLastApplied = n.stateMachine.lastApplied
	}
	if n.nextIndex > n.lastPrintNextIndex {
		Infof("Node.print %s nextIndex %d==>%d", n.address, n.lastPrintNextIndex, n.nextIndex)
		n.lastPrintNextIndex = n.nextIndex
	}
	n.printPeers()
}
func (n *Node) commit() bool {
	n.nodesMut.RLock()
	defer n.nodesMut.RUnlock()
	if n.votingsCount() == 1 {
		index := n.lastLogIndex
		if index > n.commitIndex {
			//n.storage.Sync(n.log.name)
			//n.storage.Sync(n.log.indexs.name)
			n.commitIndex = index
			//n.pipeline.commitIndex <- index
			return true
		}
		return false
	}
	var lastLogIndexs = make([]uint64, 1)
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
		//n.storage.Sync(n.log.name)
		//n.storage.Sync(n.log.indexs.name)
		n.commitIndex = index
		//n.pipeline.commitIndex <- index
		return true
	}
	return false
}
