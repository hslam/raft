// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
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
	s.node.votedFor.Reset()
	s.Start()
	return s
}

func (s *followerState) Start() {
	s.node.leader = ""
	s.node.election.Random(true)
	s.node.election.Reset()
	logger.Tracef("%s followerState.Start Term :%d", s.node.address, s.node.currentTerm.ID())
}

func (s *followerState) Update() bool {
	if s.node.commitIndex > 0 && s.node.commitIndex > s.node.stateMachine.lastApplied {
		if !atomic.CompareAndSwapInt32(&s.applying, 0, 1) {
			return true
		}
		defer atomic.StoreInt32(&s.applying, 0)
		var ch = make(chan bool, 1)
		go func(ch chan bool) {
			defer func() {
				if err := recover(); err != nil {
				}
			}()
			//var lastApplied=state.node.stateMachine.lastApplied
			s.node.log.applyCommited()
			//logger.Tracef("followerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
			ch <- true
		}(ch)
		timer := time.NewTimer(defaultCommandTimeout)
		select {
		case <-ch:
			timer.Stop()
			close(ch)
		case <-timer.C:
			logger.Tracef("%s followerState.Update applyCommited time out", s.node.address)
		}
		return true
	}
	return false
}

func (s *followerState) FixedUpdate() {
	if s.node.election.Timeout() {
		s.node.leader = ""
		s.node.votedFor.Set("")
		if !s.node.voting() {
			return
		}
		logger.Tracef("%s followerState.FixedUpdate ElectionTimeout", s.node.address)
		s.node.nextState()
		return
	}
}
func (s *followerState) String() string {
	return follower
}

func (s *followerState) StepDown() state {
	logger.Tracef("%s followerState.StepDown", s.node.address)
	s.Start()
	return s
}
func (s *followerState) NextState() state {
	if !s.node.voting() {
		return s
	}
	logger.Tracef("%s followerState.NextState", s.node.address)
	return newCandidateState(s.node)
}
