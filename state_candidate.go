// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type candidateState struct {
	node *node
}

func newCandidateState(n *node) state {
	//Tracef("%s newCandidateState",node.address)
	s := &candidateState{
		node: n,
	}
	s.Start()
	return s
}

func (s *candidateState) Start() {
	s.node.election.Random(true)
	s.node.election.Reset()
	s.node.currentTerm.Incre()
	s.node.votedFor.Set(s.node.address)
	s.node.leader = ""
	s.node.requestVotes()
	Infof("%s candidateState.Start Term :%d", s.node.address, s.node.currentTerm.ID())
}
func (s *candidateState) Update() bool {
	return false
}
func (s *candidateState) FixedUpdate() {
	if !s.node.voting() {
		s.node.lease = false
		s.node.stepDown()
		Tracef("%s candidateState.FixedUpdate non-voting", s.node.address)
		return
	} else if s.node.election.Timeout() {
		Tracef("%s candidateState.FixedUpdate ElectionTimeout", s.node.address)
		s.node.stay()
		return
	} else if s.node.votes.Count() >= s.node.Quorum() {
		Tracef("%s candidateState.FixedUpdate request Enough Votes %d Quorum %d Term %d", s.node.address, s.node.votes.Count(), s.node.Quorum(), s.node.currentTerm.ID())
		s.node.nextState()
		return
	}
}

func (s *candidateState) String() string {
	return candidate
}

func (s *candidateState) StepDown() state {
	Tracef("%s candidateState.StepDown", s.node.address)
	return newFollowerState(s.node)
}

func (s *candidateState) NextState() state {
	Tracef("%s candidateState.NextState", s.node.address)
	return newLeaderState(s.node)
}
