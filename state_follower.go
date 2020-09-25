// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"time"
)

type followerState struct {
	node *Node
	work bool
}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state := &followerState{
		node: node,
		work: true,
	}
	state.node.votedFor.Reset()
	state.Start()
	return state
}

func (state *followerState) Start() {
	state.node.leader = ""
	state.node.election.Random(true)
	state.node.election.Reset()
	Infof("%s followerState.Start Term :%d", state.node.address, state.node.currentTerm.Id())
}

func (state *followerState) Update() bool {
	if state.work {
		if state.node.commitIndex > 0 && state.node.commitIndex > state.node.stateMachine.lastApplied {
			state.work = false
			func() {
				defer func() { state.work = true }()
				var ch = make(chan bool, 1)
				go func(ch chan bool) {
					defer func() {
						if err := recover(); err != nil {
						}
					}()
					//var lastApplied=state.node.stateMachine.lastApplied
					state.node.log.applyCommited()
					//Tracef("followerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
					ch <- true
				}(ch)
				select {
				case <-ch:
					close(ch)
				case <-time.After(time.Minute):
					Tracef("%s followerState.Update applyCommited time out", state.node.address)
					Tracef("%s followerState.Update first-%d last-%d", state.node.address, state.node.firstLogIndex, state.node.lastLogIndex)
				}
			}()
			return true
		}
	}
	return false
}

func (state *followerState) FixedUpdate() {
	if state.node.election.Timeout() {
		state.node.leader = ""
		state.node.votedFor.Set("")
		if !state.node.voting() {
			return
		}
		Tracef("%s followerState.FixedUpdate ElectionTimeout", state.node.address)
		state.node.nextState()
		return
	}
}
func (state *followerState) String() string {
	return Follower
}

func (state *followerState) StepDown() State {
	Tracef("%s followerState.StepDown", state.node.address)
	state.Start()
	return state
}
func (state *followerState) NextState() State {
	if !state.node.voting() {
		return state
	}
	Tracef("%s followerState.NextState", state.node.address)
	return newCandidateState(state.node)
}
