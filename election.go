package raft

import "time"

type Election struct{
	node					*Node
	startTime 				time.Time
	electionTimeout			time.Duration
	randomElectionTimeout	time.Duration
}

func newElection(node *Node,electionTimeout time.Duration) *Election {
	election:=&Election{
		node:node,
		electionTimeout:electionTimeout,
	}
	return election
}

func (election *Election)Reset(){
	election.startTime=time.Now()
	election.randomElectionTimeout=election.electionTimeout+randomDurationTime(election.electionTimeout)
}

func (election *Election)Timeout()bool{
	if election.startTime.Add(election.randomElectionTimeout).Before(time.Now()){
		return true
	}
	return false
}