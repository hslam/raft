package raft

import "time"

type FollowerState struct{
	node							*Node
	work 							bool
}

func newFollowerState(node *Node) State {
	//Tracef("%s newFollowerState",node.address)
	state:=&FollowerState{
		node:node,
		work:true,
	}
	state.node.votedFor.Reset()
	state.Reset()
	return state
}

func (state *FollowerState)Reset(){
	state.node.leader=""
	state.node.election.Random(true)
	state.node.election.Reset()
	Debugf("%s FollowerState.Reset Term :%d",state.node.address,state.node.currentTerm.Id())
}

func (state *FollowerState) Update(){
	if state.work{
		if state.node.commitIndex.Id()>0&&state.node.commitIndex.Id()>state.node.stateMachine.lastApplied{
			state.work=false
			func() {
				var ch =make(chan bool,1)
				go func (ch chan bool) {
					defer func() {if err := recover(); err != nil {}}()
					//var lastApplied=state.node.stateMachine.lastApplied
					state.node.log.applyCommited()
					//Tracef("FollowerState.Update %s lastApplied %d==>%d",state.node.address, lastApplied,state.node.stateMachine.lastApplied)
					ch<-true
				}(ch)
				select {
				case <-ch:
					close(ch)
					state.work=true
				case <-time.After(time.Minute):
					state.work=true
					Tracef("%s FollowerState.Update applyCommited time out",state.node.address)
				}
			}()
		}
	}
}

func (state *FollowerState)FixedUpdate(){
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
	Tracef("%s FollowerState.PreState",state.node.address)
	state.Reset()
	return state
}
func (state *FollowerState)NextState()State{
	Tracef("%s FollowerState.NextState",state.node.address)
	return newCandidateState(state.node)
}