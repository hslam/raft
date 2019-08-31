package raft

import (
	"time"
	"sync"
)

type LeaderState struct{
	once sync.Once
	node					*Node
	stop					chan bool
	heartbeatTicker			*time.Ticker
}
func newLeaderState(node *Node) State {
	//Tracef("%s newLeaderState",node.address)
	state:=&LeaderState{
		node:					node,
		stop:					make(chan bool,1),
		heartbeatTicker:		time.NewTicker(node.hearbeatTick),
	}
	state.Init()
	return state
}

func (state *LeaderState)Init(){
	state.node.votedFor=""
	state.node.leader=state.node.address
	state.once.Do(func() {
		go state.run()
	})
	Debugf("%s LeaderState.Init Term :%d",state.node.address,state.node.currentTerm)

}

func (state *LeaderState) Update(){
	if state.node.AliveCount()<state.node.Quorum(){
		Tracef("%s LeaderState.Update AliveCount %d < Quorum %d",state.node.address,state.node.AliveCount(),state.node.Quorum())
		state.node.stepDown()
	}else if state.node.FollowerCount()<state.node.Quorum(){
		//Tracef("%s LeaderState.Update FollowerCount %d < Quorum %d",state.node.address,state.node.FollowerCount(),state.node.Quorum())
		//state.node.stepDown()
	}
}

func (state *LeaderState) String()string{
	return Leader
}

func (state *LeaderState)StepDown()State{
	state.stop<-true
	return newFollowerState(state.node)
}
func (state *LeaderState)NextState()State{
	state.stop<-true
	return newFollowerState(state.node)
}

func (state *LeaderState) run() {
	for{
		select {
		case <-state.stop:
			goto endfor
		case <-state.heartbeatTicker.C:
			state.node.heartbeats()
		}
	}
endfor:
}
