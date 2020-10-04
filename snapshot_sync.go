// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync/atomic"
	"time"
)

// SyncType represents a sync type.
type SyncType struct {
	Seconds int
	Changes int
}

type snapshotSync struct {
	stateMachine *stateMachine
	ticker       *time.Ticker
	syncType     *SyncType
	done         chan struct{}
	closed       int32
}

func newSnapshotSync(s *stateMachine, syncType *SyncType) *snapshotSync {
	return &snapshotSync{
		stateMachine: s,
		ticker:       time.NewTicker(time.Second * time.Duration(syncType.Seconds)),
		syncType:     syncType,
		done:         make(chan struct{}, 1),
	}
}

func (s *snapshotSync) run() {
	for {
		select {
		case <-s.ticker.C:
			changes := s.stateMachine.lastApplied - s.stateMachine.snapshotReadWriter.lastIncludedIndex.ID()
			if changes >= uint64(s.syncType.Changes) {
				s.stateMachine.SaveSnapshot()
			}
		case <-s.done:
			//logger.Tracef("snapshotSync.run Seconds-%d, Changes-%d", s.syncType.Seconds, s.syncType.Changes)
			return
		}
	}
}

func (s *snapshotSync) Stop() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.ticker.Stop()
		close(s.done)
	}
}
