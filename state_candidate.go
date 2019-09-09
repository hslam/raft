package raft

type CandidateState struct{
	node					*Node
}
func newCandidateState(node *Node) State {
	//Tracef("%s newCandidateState",node.address)
	state:=&CandidateState{
		node:node,
	}
	state.Reset()
	return state
}

func (state *CandidateState)Reset(){
	state.node.election.Reset()
	state.node.currentTerm.Incre()
	state.node.votedFor.Set(state.node.address)
	state.node.leader=""
	state.node.requestVotes()
	Debugf("%s CandidateState.Reset Term :%d",state.node.address,state.node.currentTerm.Id())
}
func (state *CandidateState) Update(){
	if state.node.election.Timeout(){
		Tracef("%s CandidateState.Update ElectionTimeout",state.node.address)
		state.node.stay()
		return
	}else if state.node.votes.Count()>=state.node.Quorum(){
		Tracef("%s CandidateState.Update request Enough Votes %d Quorum %d Term %d",state.node.address,state.node.votes.Count(),state.node.Quorum(),state.node.currentTerm.Id())
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
