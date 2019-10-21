package raft

import (
	"sync"
	"hslam.com/mgit/Mort/timer"
	"time"
)

type LeaderState struct{
	once 					*sync.Once
	node					*Node
	stop					chan bool
	notice 					chan bool
	heartbeatTicker			*time.Ticker
	checkTicker				*timer.FuncTicker
	commitTicker			*timer.FuncTicker
}
func newLeaderState(node *Node) State {
	//Tracef("%s newLeaderState",node.address)
	state:=&LeaderState{
		once:					&sync.Once{},
		node:					node,
		stop:					make(chan bool,1),
		notice:					make(chan bool,1),
		heartbeatTicker:		time.NewTicker(node.heartbeatTick),
		checkTicker:			timer.NewFuncTicker(DefaultCheckDelay,nil),
		commitTicker:			timer.NewFuncTicker(DefaultCommitDelay,nil),
	}
	state.Reset()
	go state.run()
	return state
}

func (state *LeaderState)Reset(){
	Debugf("%s LeaderState.Reset %s nextIndex:%d",state.node.address,state.node.address,state.node.nextIndex)
	if len(state.node.peers)>0{
		for _,v:=range state.node.peers{
			v.nextIndex=state.node.nextIndex
			Debugf("%s LeaderState.Reset %s nextIndex:%d",state.node.address,v.address,v.nextIndex)
		}
	}
	state.node.leader=state.node.address
	state.node.lease=true
	state.node.election.Random(false)
	state.node.election.Reset()
	Allf("%s LeaderState.Reset Term:%d",state.node.address,state.node.currentTerm.Id())
	go func(node *Node,term uint64) {
		noOperationCommand:=newNoOperationCommand()
		if ok, err := node.do(noOperationCommand,time.Minute*10);ok!=nil{
			if node.currentTerm.Id()==term{
				node.ready=true
			}
		}else if err!=nil{
			if err==ErrCommandTimeout{
				if node.currentTerm.Id()==term{
					state.node.lease=false
					state.node.stepDown()
				}
			}
		}
	}(state.node,state.node.currentTerm.Id())
}

func (state *LeaderState) Update(){
	if state.node.election.Timeout(){
		state.node.lease=false
		state.node.stepDown()
		return
	}
	if state.node.AliveCount()>=state.node.Quorum(){
		state.node.lease=true
		state.node.election.Reset()
	}
}

func (state *LeaderState) String()string{
	return Leader
}

func (state *LeaderState)StepDown()State{
	defer func() {if err := recover(); err != nil {}}()
	Tracef("%s LeaderState.StepDown",state.node.address)
	state.once.Do(func() {
		state.stop<-true
		if state.notice!=nil{
			select {
			case <-state.notice:
				close(state.notice)
			}
		}
	})
	return newFollowerState(state.node)
}
func (state *LeaderState)NextState()State{
	defer func() {if err := recover(); err != nil {}}()
	Tracef("%s LeaderState.NextState",state.node.address)
	state.once.Do(func() {
		state.stop<-true
		if state.notice!=nil{
			select {
			case <-state.notice:
				close(state.notice)
			}
		}
	})
	return newFollowerState(state.node)
}

func (state *LeaderState) run() {
	state.checkTicker.Tick(func(){
		defer func() {if err := recover(); err != nil {}}()
		state.node.check()
	})
	state.commitTicker.Tick(func() {
		defer func() {if err := recover(); err != nil {}}()
		state.node.Commit()
	})
	for {
		select {
		case <-state.heartbeatTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				state.node.heartbeats()
			}()
		case <-state.stop:
			goto endfor
		}
	}
	endfor:
	close(state.stop)
	state.heartbeatTicker.Stop()
	state.heartbeatTicker=nil
	state.checkTicker.Stop()
	state.checkTicker=nil
	state.commitTicker.Stop()
	state.commitTicker=nil
	state.notice<-true
}
