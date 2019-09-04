package raft

type FollowerState struct{
	node							*Node
}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state:=&FollowerState{
		node:node,
	}
	state.node.votedFor=""
	state.Reset()
	return state
}

func (state *FollowerState)Reset(){
	state.node.leader=""
	state.node.election.Reset()
}

func (state *FollowerState) Update(){
	if state.node.election.Timeout(){
		Tracef("%s FollowerState.Update ElectionTimeout",state.node.address)
		state.node.nextState()
		return
	}
}

func (state *FollowerState) String()string{
	return Follower
}

func (state *FollowerState)StepDown()State{
	//Tracef("%s FollowerState.PreState",state.node.address)
	state.Reset()
	return state
}
func (state *FollowerState)NextState()State{
	//Tracef("%s FollowerState.NextState",state.node.address)
	return newCandidateState(state.node)
}