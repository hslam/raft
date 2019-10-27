package raft

import (
	"time"
)

type SnapshotSync struct {
	stateMachine *StateMachine
	ticker *time.Ticker
	seconds int
	changes int
}
func newSnapshotSync(s *StateMachine ,seconds int,changes int)*SnapshotSync {
	return &SnapshotSync{
		stateMachine:s,
		ticker:time.NewTicker(time.Second*time.Duration(seconds)),
		seconds:seconds,
		changes:changes,
	}
}

func (s *SnapshotSync)run()  {
	for range s.ticker.C{
		changes:=s.stateMachine.lastApplied-s.stateMachine.snapshotReadWriter.lastIncludedIndex.Id()
		if changes>=uint64(s.changes){
			s.stateMachine.SaveSnapshot()
		}
	}
}

func (s *SnapshotSync)Stop()  {
	if s.ticker!=nil{
		s.ticker.Stop()
		s.ticker=nil
	}
}