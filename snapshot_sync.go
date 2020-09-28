// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
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
}

func newSnapshotSync(s *stateMachine, syncType *SyncType) *snapshotSync {
	return &snapshotSync{
		stateMachine: s,
		ticker:       time.NewTicker(time.Second * time.Duration(syncType.Seconds)),
		syncType:     syncType,
	}
}

func (s *snapshotSync) run() {
	for range s.ticker.C {
		changes := s.stateMachine.lastApplied - s.stateMachine.snapshotReadWriter.lastIncludedIndex.ID()
		if changes >= uint64(s.syncType.Changes) {
			s.stateMachine.SaveSnapshot()
		}
	}
}

func (s *snapshotSync) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
}
