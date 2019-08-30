package raft

import "time"

type FollowerState struct{
	node							*Node
	randomNormalOperationTimeout	time.Duration
	randomElectionTimeout			time.Duration
	startElectionTime 				time.Time

}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state:=&FollowerState{
		node:node,
	}
	state.Init()
	return state
}

func (state *FollowerState)Init(){
	state.node.votedFor=""
	state.node.leader=""
	state.startElectionTime=time.Now()
	state.randomElectionTimeout=state.node.electionTimeout+randomDurationTime(state.node.electionTimeout*DefaultRangeFactor)
	state.randomNormalOperationTimeout=state.node.normalOperationTimeout+randomDurationTime(state.node.normalOperationTimeout*DefaultRangeFactor)
	state.node.resetLastRPCTime()
	Debugf("%s FollowerState.Init Term :%d waitHearbeatTimeout :%s",state.node.address,state.node.currentTerm,state.randomNormalOperationTimeout.String())
}

func (state *FollowerState) Update(){
	if state.node.leader==""&&state.startElectionTime.Add(state.randomElectionTimeout+state.randomNormalOperationTimeout).Before(time.Now()){
		Tracef("%s FollowerState.Update ElectionTimeout",state.node.address)
		state.node.nextState()
		return
	}else if state.node.leader!=""&&state.node.lastRPCTime.Add(state.randomNormalOperationTimeout).Before(time.Now()){
		 Tracef("%s FollowerState.Update waitTimeout leader %s",state.node.address,state.node.leader)
		 state.node.nextState()
		 return
	 }
}

func (state *FollowerState) String()string{
	return Follower
}

func (state *FollowerState)StepDown()State{
	//Tracef("%s FollowerState.PreState",state.node.address)
	return state
}
func (state *FollowerState)NextState()State{
	//Tracef("%s FollowerState.NextState",state.node.address)
	return newCandidateState(state.node)
}