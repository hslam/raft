// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"runtime"
	"sync/atomic"
	"time"
)

type leaderState struct {
	node            *node
	closed          int32
	done            chan bool
	heartbeatTicker *time.Ticker
}

func newLeaderState(n *node) state {
	//logger.Tracef("%s newLeaderState",node.address)
	s := &leaderState{
		node:            n,
		done:            make(chan bool, 1),
		heartbeatTicker: time.NewTicker(n.heartbeatTick),
	}
	s.Start()
	go s.run()
	return s
}

func (s *leaderState) Start() {
	//logger.Tracef("%s leaderState.Start %s nextIndex:%d", s.node.address, s.node.address, s.node.nextIndex)
	if len(s.node.peers) > 0 {
		for _, v := range s.node.peers {
			v.nextIndex = 0
			//logger.Tracef("%s leaderState.Start %s nextIndex:%d", s.node.address, v.address, v.nextIndex)
		}
	}
	s.node.pipeline.init(s.node.lastLogIndex)
	s.node.leader = s.node.address
	s.node.lease = true
	s.node.election.Random(false)
	s.node.election.Reset()
	logger.Tracef("%s leaderState.Start Term:%d", s.node.address, s.node.currentTerm.Load())
	go func(n *node, term uint64) {
		if ok, _ := n.do(noOperationCommand, defaultCommandTimeout); ok != nil {
			if n.currentTerm.Load() == term {
				n.ready = true
				if s.node.leaderChange != nil {
					go s.node.leaderChange()
				}
				return
			}
		}
		if n.currentTerm.Load() == term {
			s.node.lease = false
			s.node.stepDown()
		}
	}(s.node, s.node.currentTerm.Load())
}

func (s *leaderState) Update() bool {
	s.node.check()
	return s.node.commit()
}

func (s *leaderState) FixedUpdate() {
	if s.node.election.Timeout() {
		s.node.lease = false
		s.node.stepDown()
		logger.Tracef("%s leaderState.FixedUpdate ElectionTimeout", s.node.address)
		return
	}
	if s.node.AliveCount() >= s.node.Quorum() {
		s.node.lease = true
		s.node.election.Reset()
	}
}

func (s *leaderState) String() string {
	return leader
}

func (s *leaderState) StepDown() state {
	logger.Tracef("%s leaderState.StepDown", s.node.address)
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.done)
	}
	return newFollowerState(s.node)
}

func (s *leaderState) NextState() state {
	logger.Tracef("%s leaderState.NextState", s.node.address)
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.done)
	}
	return newFollowerState(s.node)
}

func (s *leaderState) run() {
	for {
		runtime.Gosched()
		select {
		case <-s.heartbeatTicker.C:
			s.node.heartbeats()
		case <-s.done:
			s.heartbeatTicker.Stop()
			return
		}
	}
}
