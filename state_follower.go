package raft

import "time"

type FollowerState struct{
	server					*Server
	waitHearbeatTimeout		time.Duration
}

func newFollowerState(server *Server) State {
	Tracef("%s newFollowerState",server.address)
	state:=&FollowerState{
		server:server,
		waitHearbeatTimeout:server.waitHearbeatTimeout+randomDurationTime(server.waitHearbeatTimeout*DefaultRangeFactor),
	}
	state.server.resetLastRPCTime()
	Debugf("%s newFollowerState waitHearbeatTimeout :%s",server.address,state.waitHearbeatTimeout.String())
	return state
}
func (this *FollowerState)PreState()State{
	Tracef("%s FollowerState.PreState",this.server.address)
	return this
}
func (this *FollowerState)NextState()State{
	Tracef("%s FollowerState.NextState",this.server.address)
	return newCandidateState(this.server)
}
func (this *FollowerState)Init(){
	this.server.votedFor=""
}

func (this *FollowerState) Update(){
	if this.server.lastRPCTime.Add(this.waitHearbeatTimeout).Before(time.Now()){
		Tracef("%s FollowerState.Update Timeout",this.server.address)
		this.server.ChangeState(1)
		return
	}
}

func (this *FollowerState) String()string{
	return Follower
}