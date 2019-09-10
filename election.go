package raft

import (
	"time"
	"sync"
)

type Election struct{
	node					*Node
	once 					sync.Once
	onceDisabled 			bool
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
	if election.onceDisabled{
		election.randomElectionTimeout=election.electionTimeout+randomDurationTime(election.electionTimeout)
	}
	election.once.Do(func() {
		election.randomElectionTimeout=DefaultStartWait 
	})
}

func (election *Election)Timeout()bool{
	if election.startTime.Add(election.randomElectionTimeout).Before(time.Now()){
		election.onceDisabled=true
		return true
	}
	return false
}