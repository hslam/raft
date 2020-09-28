// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"time"
)

type followerState struct {
	node *node
	work bool
}

func newFollowerState(n *node) state {
	//Tracef("%s newFollowerState",node.address)
	s := &followerState{
		node: n,
		work: true,
	}
	s.node.votedFor.Reset()
	s.Start()
	return s
}

func (s *followerState) Start() {
	s.node.leader = ""
	s.node.election.Random(true)
	s.node.election.Reset()
	Infof("%s followerState.Start Term :%d", s.node.address, s.node.currentTerm.ID())
}

func (s *followerState) Update() bool {
	if s.work {
		if s.node.commitIndex > 0 && s.node.commitIndex > s.node.stateMachine.lastApplied {
			s.work = false
			func() {
				defer func() { s.work = true }()
				var ch = make(chan bool, 1)
				go func(ch chan bool) {
					defer func() {
						if err := recover(); err != nil {
						}
					}()
					//var lastApplied=state.node.stateMachine.lastApplied
					s.node.log.applyCommited()
					//Tracef("followerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
					ch <- true
				}(ch)
				select {
				case <-ch:
					close(ch)
				case <-time.After(time.Minute):
					Tracef("%s followerState.Update applyCommited time out", s.node.address)
					Tracef("%s followerState.Update first-%d last-%d", s.node.address, s.node.firstLogIndex, s.node.lastLogIndex)
				}
			}()
			return true
		}
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
		Tracef("%s followerState.FixedUpdate ElectionTimeout", s.node.address)
		s.node.nextState()
		return
	}
}
func (s *followerState) String() string {
	return follower
}

func (s *followerState) StepDown() state {
	Tracef("%s followerState.StepDown", s.node.address)
	s.Start()
	return s
}
func (s *followerState) NextState() state {
	if !s.node.voting() {
		return s
	}
	Tracef("%s followerState.NextState", s.node.address)
	return newCandidateState(s.node)
}
