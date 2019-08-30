package raft

import "time"

type CandidateState struct{
	server					*Server
	startElectionTime 		time.Time
}
func newCandidateState(server *Server) State {
	Tracef("%s newCandidateState",server.address)
	state:=&CandidateState{
		server:server,
		startElectionTime:time.Now(),
	}
	state.Init()
	return state
}
func (this *CandidateState)PreState()State{
	Tracef("%s CandidateState.PreState",this.server.address)
	return newFollowerState(this.server)
}

func (this *CandidateState)NextState()State{
	Tracef("%s CandidateState.NextState",this.server.address)
	return newLeaderState(this.server)
}

func (this *CandidateState)Init(){
	this.server.votedFor=""
	this.server.leader=""
	this.server.requestVotes()
}
func (this *CandidateState) Update(){
	if this.server.voteCount>=this.server.Quorum(){
		Tracef("%s CandidateState.Update request Enough Votes %d Quorum %d",this.server.address,this.server.voteCount,this.server.Quorum())
		this.server.ChangeState(1)
		return
	}else if this.server.voteTotal>=this.server.NodesLen(){
		Tracef("%s CandidateState.Update request All Votes",this.server.address)
		this.server.ChangeState(-1)
	}else if this.startElectionTime.Add(this.server.electionTimeout).Before(time.Now()){
		Tracef("%s CandidateState.Update ElectionTimeout",this.server.address)
		this.server.ChangeState(-1)
		return
	}
}

func (this *CandidateState) String()string{
	return Candidate
}