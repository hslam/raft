// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"runtime"
	"sync/atomic"
	"time"
)

type followerState struct {
	node     *node
	applying int32
}

func newFollowerState(n *node) state {
	//logger.Tracef("%s newFollowerState",node.address)
	s := &followerState{
		node: n,
	}
	s.node.votedFor.Store("")
	s.Start()
	return s
}

func (s *followerState) Start() {
	s.node.leader.Store("")
	s.node.election.Random(true)
	s.node.election.Reset()
	s.node.logger.Tracef("%s followerState.Start Term :%d", s.node.address, s.node.currentTerm.Load())
}

func (s *followerState) Update() bool {
	if s.node.commitIndex.ID() > 0 && s.node.commitIndex.ID() > s.node.stateMachine.lastApplied {
		if atomic.CompareAndSwapInt32(&s.applying, 0, 1) {
			defer atomic.StoreInt32(&s.applying, 0)
			var ch = make(chan bool, 1)
			go func(ch chan bool) {
				//var lastApplied=state.node.stateMachine.lastApplied
				s.node.log.applyCommited()
				//s.node.logger.Tracef("followerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
				ch <- true
			}(ch)
			timer := time.NewTimer(defaultCommandTimeout)
			runtime.Gosched()
			select {
			case <-ch:
				timer.Stop()
			case <-timer.C:
				//s.node.logger.Tracef("%s followerState.Update applyCommited time out", s.node.address)
			}
		}
		return true
	}
	return false
}

func (s *followerState) FixedUpdate() {
	if s.node.election.Timeout() && s.node.lastLogIndex >= s.node.commitIndex.ID() {
		s.node.leader.Store("")
		s.node.votedFor.Store("")
		s.node.logger.Tracef("%s followerState.FixedUpdate ElectionTimeout", s.node.address)
		s.node.nextState()
	}
}

func (s *followerState) String() string {
	return follower
}

func (s *followerState) StepDown() state {
	s.node.logger.Tracef("%s followerState.StepDown", s.node.address)
	s.Start()
	return s
}

func (s *followerState) NextState() state {
	if !s.node.voting() {
		return s
	}
	s.node.logger.Tracef("%s followerState.NextState", s.node.address)
	return newCandidateState(s.node)
}
