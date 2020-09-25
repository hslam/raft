// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type candidateState struct {
	node *Node
}

func newCandidateState(node *Node) State {
	//Tracef("%s newCandidateState",node.address)
	state := &candidateState{
		node: node,
	}
	state.Start()
	return state
}

func (state *candidateState) Start() {
	state.node.election.Random(true)
	state.node.election.Reset()
	state.node.currentTerm.Incre()
	state.node.votedFor.Set(state.node.address)
	state.node.leader = ""
	state.node.requestVotes()
	Infof("%s candidateState.Start Term :%d", state.node.address, state.node.currentTerm.Id())
}
func (state *candidateState) Update() bool {
	return false
}
func (state *candidateState) FixedUpdate() {
	if !state.node.voting() {
		state.node.lease = false
		state.node.stepDown()
		Tracef("%s candidateState.FixedUpdate non-voting", state.node.address)
		return
	} else if state.node.election.Timeout() {
		Tracef("%s candidateState.FixedUpdate ElectionTimeout", state.node.address)
		state.node.stay()
		return
	} else if state.node.votes.Count() >= state.node.Quorum() {
		Tracef("%s candidateState.FixedUpdate request Enough Votes %d Quorum %d Term %d", state.node.address, state.node.votes.Count(), state.node.Quorum(), state.node.currentTerm.Id())
		state.node.nextState()
		return
	}
}

func (state *candidateState) String() string {
	return Candidate
}

func (state *candidateState) StepDown() State {
	Tracef("%s candidateState.StepDown", state.node.address)
	return newFollowerState(state.node)
}

func (state *candidateState) NextState() State {
	Tracef("%s candidateState.NextState", state.node.address)
	return newLeaderState(state.node)
}
