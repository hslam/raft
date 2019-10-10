package raft

import (
	"time"
	"sync"
)

type Election struct{
	node					*Node
	once 					sync.Once
	onceDisabled 			bool
	random 					bool
	startTime 				time.Time
	defaultElectionTimeout	time.Duration
	electionTimeout			time.Duration
}

func newElection(node *Node,electionTimeout time.Duration) *Election {
	election:=&Election{
		node:node,
		defaultElectionTimeout:electionTimeout,
		random:true,
	}
	return election
}

func (election *Election)Reset(){
	election.startTime=time.Now()
	if election.onceDisabled{
		if election.random{
			election.electionTimeout=election.defaultElectionTimeout+randomDurationTime(election.defaultElectionTimeout)
		}else{
			election.electionTimeout=election.defaultElectionTimeout
		}
	}
	election.once.Do(func() {
		election.electionTimeout=DefaultStartWait
	})
}
func (election *Election)Random(random bool){
	election.random=random
}
func (election *Election)Timeout()bool{
	if election.startTime.Add(election.electionTimeout).Before(time.Now()){
		election.onceDisabled=true
		return true
	}
	return false
}