package raft

import (
	"hslam.com/mgit/Mort/timer"
	"sync"
)

type LeaderState struct{
	once 					sync.Once
	node					*Node
	stop					chan bool
	heartbeatTicker			*timer.Ticker
	checkTicker				*timer.Ticker
}
func newLeaderState(node *Node) State {
	//Tracef("%s newLeaderState",node.address)
	state:=&LeaderState{
		node:					node,
		stop:					make(chan bool,1),
		heartbeatTicker:		timer.NewFuncTicker(node.hearbeatTick),
		checkTicker:			timer.NewFuncTicker(DefaultCheckDelay),
	}
	state.Reset()
	go state.run()
	return state
}

func (state *LeaderState)Reset(){
	if len(state.node.peers)>0{
		for _,v:=range state.node.peers{
			v.nextIndex=state.node.lastLogIndex.Id()+1
			Debugf("%s LeaderState.Reset %s nextIndex :%d",state.node.address,v.address,v.nextIndex)
		}
	}
	state.node.leader=state.node.address
	Allf("%s LeaderState.Reset Term :%d",state.node.address,state.node.currentTerm.Id())
}

func (state *LeaderState) Update(){
	//if state.node.AliveCount()<state.node.Quorum(){
	//	Tracef("%s LeaderState.Update AliveCount %d < Quorum %d",state.node.address,state.node.AliveCount(),state.node.Quorum())
	//	state.node.stepDown()
	//}
	state.node.Commit()
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
	state.checkTicker.Tick(func() {
		state.node.check()
	})
	state.heartbeatTicker.Tick(func() {
		state.node.heartbeats()
	})
	for{
		select {
		case <-state.stop:
			goto endfor
		}
	}
endfor:
	close(state.stop)
	state.heartbeatTicker.Stop()
	state.checkTicker.Stop()
}
