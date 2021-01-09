// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type candidateState struct {
	node *node
}

func newCandidateState(n *node) state {
	//logger.Tracef("%s newCandidateState",node.address)
	s := &candidateState{
		node: n,
	}
	s.Start()
	return s
}

func (s *candidateState) Start() {
	s.node.election.Random(true)
	s.node.election.Reset()
	s.node.currentTerm.Add(1)
	s.node.votedFor.Store(s.node.address)
	s.node.leader.Store("")
	s.node.requestVotes()
	logger.Tracef("%s candidateState.Start Term :%d", s.node.address, s.node.currentTerm.Load())
}

func (s *candidateState) Update() bool {
	return false
}

func (s *candidateState) FixedUpdate() {
	if s.node.election.Timeout() {
		logger.Tracef("%s candidateState.FixedUpdate ElectionTimeout Votes %d Quorum %d Term %d", s.node.address, s.node.votes.Count(), s.node.Quorum(), s.node.currentTerm.Load())
		s.node.stay()
	} else if s.node.votes.Count() >= s.node.Quorum() {
		logger.Tracef("%s candidateState.FixedUpdate request Enough Votes %d Quorum %d Term %d", s.node.address, s.node.votes.Count(), s.node.Quorum(), s.node.currentTerm.Load())
		s.node.nextState()
	}
}

func (s *candidateState) String() string {
	return candidate
}

func (s *candidateState) StepDown() state {
	logger.Tracef("%s candidateState.StepDown", s.node.address)
	return newFollowerState(s.node)
}

func (s *candidateState) NextState() state {
	logger.Tracef("%s candidateState.NextState", s.node.address)
	return newLeaderState(s.node)
}
