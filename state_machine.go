// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var errOrder = errors.New("Order Error")
var errRepeated = errors.New("This command had repeated executed")

type stateMachine struct {
	mu                 sync.RWMutex
	node               *node
	lastApplied        uint64
	configuration      *configuration
	snapshot           Snapshot
	snapshotReadWriter *snapshotReadWriter
	snapshotPolicy     SnapshotPolicy
	snapshotSyncs      []*snapshotSync
	saves              []*SyncType
	saveLog            bool
	always             bool
}

func newStateMachine(n *node) *stateMachine {
	s := &stateMachine{
		node:               n,
		configuration:      newConfiguration(n),
		snapshotReadWriter: newSnapshotReadWriter(n, defaultSnapshot, false),
		saves:              []*SyncType{},
	}
	s.SetSnapshotPolicy(DefalutSync)
	return s
}

func (s *stateMachine) Apply(index uint64, command Command) (reply interface{}, err error, applyErr error) {
	s.mu.Lock()
	reply, err, applyErr = s.apply(index, command)
	s.mu.Unlock()
	return
}

func (s *stateMachine) Lock() {
	s.mu.Lock()
}

func (s *stateMachine) Unlock() {
	s.mu.Unlock()
}

func (s *stateMachine) apply(index uint64, command Command) (reply interface{}, err error, applyErr error) {
	if index <= s.lastApplied {
		return nil, nil, errRepeated
	}
	if index != s.lastApplied+1 {
		return nil, nil, errOrder
	}
	if command.Type() > 0 {
		reply, err = command.Do(s.node.context)
	} else {
		reply, err = command.Do(s.node)
	}
	s.lastApplied = index
	if s.always {
		s.saveSnapshot()
	}
	return reply, err, nil
}

func (s *stateMachine) SetSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	s.mu.Lock()
	s.setSnapshotPolicy(snapshotPolicy)
	s.mu.Unlock()
}

func (s *stateMachine) setSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	s.snapshotPolicy = snapshotPolicy
	s.always = false
	s.StopSnapshotSyncs()
	switch s.snapshotPolicy {
	case Never:
	case EverySecond:
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{1, 1}))
		go s.run()
	case EveryMinute:
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{60, 1}))
		go s.run()
	case EveryHour:
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{3600, 1}))
		go s.run()
	case EveryDay:
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{86400, 1}))
		go s.run()
	case DefalutSync:
		s.saves = []*SyncType{
			{secondsSaveSnapshot1, changesSaveSnapshot1},
			{secondsSaveSnapshot2, changesSaveSnapshot2},
			{secondsSaveSnapshot3, changesSaveSnapshot3},
		}
		for _, v := range s.saves {
			s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, v))
		}
		go s.run()
	case CustomSync:
		for _, v := range s.saves {
			s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, v))
		}
		go s.run()
	case Always:
		s.always = true
	}
}

func (s *stateMachine) SetSyncTypes(saves []*SyncType) {
	s.mu.Lock()
	s.saves = saves
	s.setSnapshotPolicy(CustomSync)
	s.mu.Unlock()
}

func (s *stateMachine) SetSnapshot(snapshot Snapshot) {
	s.mu.Lock()
	s.snapshot = snapshot
	s.mu.Unlock()
}

func (s *stateMachine) SaveSnapshot() error {
	s.mu.Lock()
	err := s.saveSnapshot()
	s.mu.Unlock()
	return err
}

func (s *stateMachine) saveSnapshot() error {
	if s.snapshot != nil {
		if s.lastApplied > s.snapshotReadWriter.lastIncludedIndex.ID() &&
			atomic.CompareAndSwapInt32(&s.snapshotReadWriter.work, 0, 1) {
			defer atomic.StoreInt32(&s.snapshotReadWriter.work, 0)
			s.node.logger.Tracef("stateMachine.saveSnapshot %s Start", s.node.address)
			var lastPrintLastIncludedIndex = s.snapshotReadWriter.lastIncludedIndex.ID()
			var lastIncludedIndex, lastIncludedTerm uint64
			if s.node.lastLogIndex == s.lastApplied {
				lastIncludedIndex = s.node.lastLogIndex
				lastIncludedTerm = s.node.lastLogTerm
			} else {
				lastIncludedIndex = s.lastApplied
				entry := s.node.log.read(lastIncludedIndex)
				if entry == nil {
					return errors.New("this entry is not existed")
				}
				lastIncludedTerm = entry.Term
			}
			s.snapshotReadWriter.clearFlush()
			_, err := s.snapshot.Save(s.snapshotReadWriter)
			startTime := time.Now()
			s.snapshotReadWriter.Rename()
			s.snapshotReadWriter.Reset(lastIncludedIndex, lastIncludedTerm)
			if s.node.isLeader() {
				s.node.log.deleteBefore(minUint64(s.node.minNextIndex(), lastIncludedIndex))
				select {
				case s.snapshotReadWriter.trigger <- struct{}{}:
				default:
				}
			} else {
				s.node.log.deleteBefore(lastIncludedIndex)
			}
			duration := time.Now().Sub(startTime)
			s.node.logger.Tracef("stateMachine.saveSnapshot %s Finish lastIncludedIndex %d==>%d duration:%v", s.node.address, lastPrintLastIncludedIndex, s.snapshotReadWriter.lastIncludedIndex.ID(), duration)
			return err
		}
		return nil
	}
	return ErrSnapshotCodecNil
}

func (s *stateMachine) RecoverSnapshot() error {
	if s.snapshot != nil {
		s.snapshotReadWriter.load()
		_, err := s.snapshot.Recover(s.snapshotReadWriter)
		return err
	}
	return ErrSnapshotCodecNil
}

func (s *stateMachine) recover() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.node.storage.IsEmpty(defaultSnapshot) {
		if !s.node.storage.IsEmpty(defaultSnapshot + defaultTmp) {
			s.snapshotReadWriter.Rename()
		} else {
			return errors.New(defaultSnapshot + " file is empty")
		}
	}
	if s.snapshotReadWriter.lastIncludedIndex.ID() <= s.lastApplied && s.lastApplied > 0 && s.snapshotReadWriter.lastIncludedIndex.ID() > 0 {
		return nil
	}
	err := s.RecoverSnapshot()
	if err != nil {
		return err
	}
	if s.node.lastLogIndex < s.snapshotReadWriter.lastIncludedIndex.ID() {
		var lastLogIndex = s.node.lastLogIndex
		s.node.lastLogIndex = s.snapshotReadWriter.lastIncludedIndex.ID()
		s.node.logger.Tracef("stateMachine.recover %s lastLogIndex %d==>%d", s.node.address, lastLogIndex, s.node.lastLogIndex)
	}
	if s.node.lastLogTerm < s.snapshotReadWriter.lastIncludedTerm.ID() {
		var lastLogTerm = s.node.lastLogTerm
		s.node.lastLogTerm = s.snapshotReadWriter.lastIncludedTerm.ID()
		s.node.logger.Tracef("stateMachine.recover %s lastLogTerm %d==>%d", s.node.address, lastLogTerm, s.node.lastLogTerm)
	}
	if s.lastApplied < s.snapshotReadWriter.lastIncludedIndex.ID() {
		var lastApplied = s.lastApplied
		s.lastApplied = s.snapshotReadWriter.lastIncludedIndex.ID()
		s.node.logger.Tracef("stateMachine.recover %s lastApplied %d==>%d", s.node.address, lastApplied, s.lastApplied)
	}
	if s.node.commitIndex.ID() < s.snapshotReadWriter.lastIncludedIndex.ID() {
		var commitIndex = s.node.commitIndex.ID()
		s.node.commitIndex.Set(s.snapshotReadWriter.lastIncludedIndex.ID())
		s.node.logger.Tracef("stateMachine.recover %s commitIndex %d==>%d", s.node.address, commitIndex, s.node.commitIndex.ID())
	}
	if s.node.nextIndex < s.snapshotReadWriter.lastIncludedIndex.ID()+1 {
		var nextIndex = s.node.nextIndex
		s.node.nextIndex = s.snapshotReadWriter.lastIncludedIndex.ID() + 1
		s.node.logger.Tracef("stateMachine.recover %s nextIndex %d==>%d", s.node.address, nextIndex, s.node.nextIndex)
		s.node.log.wal.Reset()
		s.node.log.wal.InitFirstIndex(s.node.nextIndex)
		s.node.log.load()
	}
	return nil
}

func (s *stateMachine) Reset() {
	s.lastApplied = 0
	s.snapshotReadWriter.Reset(0, 0)
	s.snapshotReadWriter.lastTarIndex.Set(0)
	s.configuration.reset()
}

func (s *stateMachine) load() {
	s.snapshotReadWriter.load()
	s.configuration.load()
	s.configuration.reconfiguration()
}

func (s *stateMachine) append(offset uint64, p []byte) {
	s.snapshotReadWriter.Append(offset, p)
}

func (s *stateMachine) run() {
	s.node.logger.Tracef("stateMachine.run SnapshotSyncs %d", len(s.snapshotSyncs))
	for _, snapshotSync := range s.snapshotSyncs {
		go snapshotSync.run()
	}
}

func (s *stateMachine) StopSnapshotSyncs() {
	//s.node.logger.Tracef("stateMachine.Stop %d", len(s.snapshotSyncs))
	for _, snapshotSync := range s.snapshotSyncs {
		if snapshotSync != nil {
			snapshotSync.Stop()
			snapshotSync = nil
		}
	}
	s.snapshotSyncs = make([]*snapshotSync, 0)
}

func (s *stateMachine) Stop() {
	s.StopSnapshotSyncs()
	s.snapshotReadWriter.Stop()
}
