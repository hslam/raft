package raft

import (
	"time"
)

type FollowerState struct {
	node *Node
	work bool
}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state := &FollowerState{
		node: node,
		work: true,
	}
	state.node.votedFor.Reset()
	state.Start()
	return state
}

func (state *FollowerState) Start() {
	state.node.leader = ""
	state.node.election.Random(true)
	state.node.election.Reset()
	Infof("%s FollowerState.Start Term :%d", state.node.address, state.node.currentTerm.Id())
}

func (state *FollowerState) Update() bool {
	if state.work {
		if state.node.commitIndex.Id() > 0 && state.node.commitIndex.Id() > state.node.stateMachine.lastApplied {
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
					//Tracef("FollowerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
					ch <- true
				}(ch)
				select {
				case <-ch:
					close(ch)
				case <-time.After(time.Minute):
					Tracef("%s FollowerState.Update applyCommited time out", state.node.address)
					Tracef("%s FollowerState.Update first-%d last-%d", state.node.address, state.node.firstLogIndex, state.node.lastLogIndex)
				}
			}()
			return true
		}
	}
	return false
}

func (state *FollowerState) FixedUpdate() {
	if state.node.election.Timeout() {
		state.node.leader = ""
		state.node.votedFor.Set("")
		if !state.node.voting() {
			return
		}
		Tracef("%s FollowerState.FixedUpdate ElectionTimeout", state.node.address)
		state.node.nextState()
		return
	}
}
func (state *FollowerState) String() string {
	return Follower
}

func (state *FollowerState) StepDown() State {
	Tracef("%s FollowerState.StepDown", state.node.address)
	state.Start()
	return state
}
func (state *FollowerState) NextState() State {
	if !state.node.voting() {
		return state
	}
	Tracef("%s FollowerState.NextState", state.node.address)
	return newCandidateState(state.node)
}
