package raft

import "time"

type LeaderState struct{
	server					*Server
	stop					chan bool
	heartbeatTicker			*time.Ticker
}
func newLeaderState(server *Server) State {
	Tracef("%s newLeaderState",server.address)
	state:=&LeaderState{
		server:					server,
		stop:					make(chan bool,1),
		heartbeatTicker:		time.NewTicker(server.hearbeatTick),
	}
	state.Init()
	go state.run()
	return state
}

func (this *LeaderState)PreState()State{
	this.stop<-true
	return newFollowerState(this.server)
}
func (this *LeaderState)NextState()State{
	this.stop<-true
	return newFollowerState(this.server)
}
func (this *LeaderState)Init(){
	this.server.currentTerm+=1
	this.server.leader=this.server.address
	this.server.votedFor=this.server.address
}

func (this *LeaderState) Update(){
}
func (this *LeaderState) String()string{
	return Leader
}

func (this *LeaderState) run() {
	for{
		select {
		case <-this.stop:
			goto endfor
		case <-this.heartbeatTicker.C:
			this.server.heartbeats()
		}
	}
endfor:
}