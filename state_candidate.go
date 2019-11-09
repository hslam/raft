package raft

type CandidateState struct{
	node					*Node
}
func newCandidateState(node *Node) State {
	//Tracef("%s newCandidateState",node.address)
	state:=&CandidateState{
		node:node,
	}
	state.Start()
	return state
}

func (state *CandidateState)Start(){
	state.node.election.Random(true)
	state.node.election.Reset()
	state.node.currentTerm.Incre()
	state.node.votedFor.Set(state.node.address)
	state.node.leader=""
	state.node.requestVotes()
	Debugf("%s CandidateState.Reset Term :%d",state.node.address,state.node.currentTerm.Id())
}
func (state *CandidateState) Update()bool{
	return false
}
func (state *CandidateState)FixedUpdate(){
	if state.node.election.Timeout(){
		Tracef("%s CandidateState.Update ElectionTimeout",state.node.address)
		state.node.stay()
		return
	}else if state.node.votes.Count()>=state.node.Quorum(){
		Tracef("%s CandidateState.Update request Enough Votes %d Quorum %d Term %d",state.node.address,state.node.votes.Count(),state.node.Quorum(),state.node.currentTerm.Id())
		state.node.nextState()
		return
	}
}

func (state *CandidateState) String()string{
	return Candidate
}

func (state *CandidateState)StepDown()State{
	Tracef("%s CandidateState.PreState",state.node.address)
	return newFollowerState(state.node)
}

func (state *CandidateState)NextState()State{
	Tracef("%s CandidateState.NextState",state.node.address)
	return newLeaderState(state.node)
}
