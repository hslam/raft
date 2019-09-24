package raft

type FollowerState struct{
	node							*Node
}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state:=&FollowerState{
		node:node,
	}
	state.node.votedFor.Reset()
	state.Reset()
	return state
}

func (state *FollowerState)Reset(){
	state.node.leader=""
	state.node.election.Reset()
	Debugf("%s FollowerState.Reset Term :%d",state.node.address,state.node.currentTerm.Id())
}

func (state *FollowerState) Update(){
	if state.node.election.Timeout(){
		Tracef("%s FollowerState.Update ElectionTimeout",state.node.address)
		state.node.nextState()
		return
	}
	if state.node.commitIndex.Id()>0&&state.node.commitIndex.Id()>state.node.stateMachine.lastApplied{
		//var lastApplied=state.node.stateMachine.lastApplied
		state.node.log.applyCommited()
		//Tracef("FollowerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
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