// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"time"
)

type leaderState struct {
	once            *sync.Once
	node            *node
	stop            chan bool
	notice          chan bool
	heartbeatTicker *time.Ticker
}

func newLeaderState(n *node) state {
	//logger.Tracef("%s newLeaderState",node.address)
	s := &leaderState{
		once:            &sync.Once{},
		node:            n,
		stop:            make(chan bool, 1),
		notice:          make(chan bool, 1),
		heartbeatTicker: time.NewTicker(n.heartbeatTick),
	}
	s.Start()
	go s.run()
	return s
}

func (s *leaderState) Start() {
	//logger.Tracef("%s leaderState.Start %s nextIndex:%d", s.node.address, s.node.address, s.node.nextIndex)
	if len(s.node.peers) > 0 {
		for _, v := range s.node.peers {
			v.nextIndex = 0
			//logger.Tracef("%s leaderState.Start %s nextIndex:%d", s.node.address, v.address, v.nextIndex)
		}
	}
	s.node.pipeline.init(s.node.lastLogIndex)
	s.node.leader = s.node.address
	s.node.lease = true
	s.node.election.Random(false)
	s.node.election.Reset()
	logger.Tracef("%s leaderState.Start Term:%d", s.node.address, s.node.currentTerm.Load())
	go func(n *node, term uint64) {
		noOperationCommand := NewNoOperationCommand()
		if ok, _ := n.do(noOperationCommand, defaultCommandTimeout*10); ok != nil {
			if n.currentTerm.Load() == term {
				n.ready = true
				return
			}
		}
		if n.currentTerm.Load() == term {
			s.node.lease = false
			s.node.stepDown()
		}
	}(s.node, s.node.currentTerm.Load())
}

func (s *leaderState) Update() bool {
	s.node.check()
	return s.node.commit()
}
func (s *leaderState) FixedUpdate() {
	if !s.node.voting() {
		s.node.lease = false
		s.node.stepDown()
		logger.Tracef("%s leaderState.FixedUpdate non-voting", s.node.address)
		return
	} else if s.node.election.Timeout() {
		s.node.lease = false
		s.node.stepDown()
		logger.Tracef("%s leaderState.FixedUpdate ElectionTimeout", s.node.address)
		return
	}
	if s.node.AliveCount() >= s.node.Quorum() {
		s.node.lease = true
		s.node.election.Reset()
	}
}

func (s *leaderState) String() string {
	return leader
}

func (s *leaderState) StepDown() state {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	logger.Tracef("%s leaderState.StepDown", s.node.address)
	s.once.Do(func() {
		s.stop <- true
		if s.notice != nil {
			select {
			case <-s.notice:
				close(s.notice)
			}
		}
	})
	return newFollowerState(s.node)
}
func (s *leaderState) NextState() state {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	logger.Tracef("%s leaderState.NextState", s.node.address)
	s.once.Do(func() {
		s.stop <- true
		if s.notice != nil {
			select {
			case <-s.notice:
				close(s.notice)
			}
		}
	})
	return newFollowerState(s.node)
}

func (s *leaderState) run() {
	for {
		select {
		case <-s.heartbeatTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				s.node.heartbeats()
			}()
		case <-s.stop:
			goto endfor
		}
	}
endfor:
	close(s.stop)
	s.heartbeatTicker.Stop()
	s.heartbeatTicker = nil
	s.notice <- true
}
