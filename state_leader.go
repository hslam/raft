// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"time"
)

type leaderState struct {
	once            *sync.Once
	node            *Node
	stop            chan bool
	notice          chan bool
	heartbeatTicker *time.Ticker
}

func newLeaderState(node *Node) State {
	//Tracef("%s newLeaderState",node.address)
	state := &leaderState{
		once:            &sync.Once{},
		node:            node,
		stop:            make(chan bool, 1),
		notice:          make(chan bool, 1),
		heartbeatTicker: time.NewTicker(node.heartbeatTick),
	}
	state.Start()
	go state.run()
	return state
}

func (state *leaderState) Start() {
	Debugf("%s leaderState.Start %s nextIndex:%d", state.node.address, state.node.address, state.node.nextIndex)
	if len(state.node.peers) > 0 {
		for _, v := range state.node.peers {
			v.nextIndex = 0
			Debugf("%s leaderState.Start %s nextIndex:%d", state.node.address, v.address, v.nextIndex)
		}
	}
	state.node.pipeline.init(state.node.lastLogIndex)
	state.node.leader = state.node.address
	state.node.lease = true
	state.node.election.Random(false)
	state.node.election.Reset()
	Infof("%s leaderState.Start Term:%d", state.node.address, state.node.currentTerm.Id())
	go func(node *Node, term uint64) {
		noOperationCommand := NewNoOperationCommand()
		if ok, _ := node.do(noOperationCommand, time.Minute*10); ok != nil {
			if node.currentTerm.Id() == term {
				node.ready = true
				return
			}
		}
		if node.currentTerm.Id() == term {
			state.node.lease = false
			state.node.stepDown()
		}
	}(state.node, state.node.currentTerm.Id())
}

func (state *leaderState) Update() bool {
	state.node.check()
	return state.node.commit()
}
func (state *leaderState) FixedUpdate() {
	if !state.node.voting() {
		state.node.lease = false
		state.node.stepDown()
		Tracef("%s leaderState.FixedUpdate non-voting", state.node.address)
		return
	} else if state.node.election.Timeout() {
		state.node.lease = false
		state.node.stepDown()
		Tracef("%s leaderState.FixedUpdate ElectionTimeout", state.node.address)
		return
	}
	if state.node.AliveCount() >= state.node.Quorum() {
		state.node.lease = true
		state.node.election.Reset()
	}
}

func (state *leaderState) String() string {
	return Leader
}

func (state *leaderState) StepDown() State {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	Tracef("%s leaderState.StepDown", state.node.address)
	state.once.Do(func() {
		state.stop <- true
		if state.notice != nil {
			select {
			case <-state.notice:
				close(state.notice)
			}
		}
	})
	return newFollowerState(state.node)
}
func (state *leaderState) NextState() State {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	Tracef("%s leaderState.NextState", state.node.address)
	state.once.Do(func() {
		state.stop <- true
		if state.notice != nil {
			select {
			case <-state.notice:
				close(state.notice)
			}
		}
	})
	return newFollowerState(state.node)
}

func (state *leaderState) run() {
	for {
		select {
		case <-state.heartbeatTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				state.node.heartbeats()
			}()
		case <-state.stop:
			goto endfor
		}
	}
endfor:
	close(state.stop)
	state.heartbeatTicker.Stop()
	state.heartbeatTicker = nil
	state.notice <- true
}
