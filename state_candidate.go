package raft

import "time"

type CandidateState struct{
	node					*Node
	startElectionTime 		time.Time
	randomElectionTimeout		time.Duration
}
func newCandidateState(node *Node) State {
	//Tracef("%s newCandidateState",node.address)
	state:=&CandidateState{
		node:node,
	}
	state.Init()
	return state
}

func (state *CandidateState)Init(){
	state.startElectionTime=time.Now()
	state.randomElectionTimeout=state.node.electionTimeout+randomDurationTime(state.node.electionTimeout*DefaultRangeFactor)
	state.node.currentTerm+=1
	state.node.votedFor=state.node.address
	state.node.leader=""
	state.node.requestVotes()
	Debugf("%s CandidateState.Init Term :%d",state.node.address,state.node.currentTerm)
}
func (state *CandidateState) Update(){
	if state.startElectionTime.Add(state.randomElectionTimeout).Before(time.Now()){
		Tracef("%s CandidateState.Update ElectionTimeout",state.node.address)
		state.node.stay()
		return
	}else if state.node.votes.Count()>=state.node.Quorum(){
		Tracef("%s CandidateState.Update request Enough Votes %d Quorum %d Term %d",state.node.address,state.node.votes.Count(),state.node.Quorum(),state.node.currentTerm)
		state.node.nextState()
		return
	}else if state.node.votes.Total()>=state.node.NodesCount(){
		Tracef("%s CandidateState.Update request All Replies",state.node.address)
		state.node.stepDown()
	}else if state.node.votes.Total()>=maxInt(state.node.NodesCount()-1,state.node.Quorum()){
		Tracef("%s CandidateState.Update request All-1 Replies",state.node.address)
		state.node.stepDown()
	}
}

func (state *CandidateState) String()string{
	return Candidate
}

func (state *CandidateState)StepDown()State{
	//Tracef("%s CandidateState.PreState",state.node.address)
	return newFollowerState(state.node)
}

func (state *CandidateState)NextState()State{
	//Tracef("%s CandidateState.NextState",state.node.address)
	return newLeaderState(state.node)
}
