// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"sync"
	"time"
)

var errOrder = errors.New("Order Error")
var errRepeated = errors.New("This command had repeated executed")

type StateMachine struct {
	mu                 sync.RWMutex
	node               *Node
	lastApplied        uint64
	configuration      *Configuration
	snapshot           Snapshot
	snapshotReadWriter *SnapshotReadWriter
	snapshotPolicy     SnapshotPolicy
	snapshotSyncs      []*snapshotSync
	saves              []*SyncType
	saveLog            bool
	always             bool
}

func newStateMachine(node *Node) *StateMachine {
	s := &StateMachine{
		node:               node,
		configuration:      newConfiguration(node),
		snapshotReadWriter: newSnapshotReadWriter(node, DefaultSnapshot, false),
		saves:              []*SyncType{},
	}
	s.SetSnapshotPolicy(DefalutSync)
	return s
}
func (s *StateMachine) Apply(index uint64, command Command) (reply interface{}, err error, applyErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.apply(index, command)
}
func (s *StateMachine) Lock() {
	s.mu.Lock()
}
func (s *StateMachine) Unlock() {
	s.mu.Unlock()
}
func (s *StateMachine) apply(index uint64, command Command) (reply interface{}, err error, applyErr error) {
	if index <= s.lastApplied {
		return nil, nil, errRepeated
	} else if index != s.lastApplied+1 {
		return nil, nil, errOrder
	}
	defer func() {
		s.lastApplied = index
		if s.always {
			s.saveSnapshot()
		}
	}()
	if command.Type() >= 0 {
		reply, err = command.Do(s.node.context)
	} else {
		reply, err = command.Do(s.node)
	}
	return reply, err, nil
}

func (s *StateMachine) SetSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setSnapshotPolicy(snapshotPolicy)
}

func (s *StateMachine) setSnapshotPolicy(snapshotPolicy SnapshotPolicy) {
	s.snapshotPolicy = snapshotPolicy
	s.always = false
	switch s.snapshotPolicy {
	case Never:
		s.Stop()
	case EverySecond:
		s.Stop()
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{1, 1}))
		go s.run()
	case EveryMinute:
		s.Stop()
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{60, 1}))
		go s.run()
	case EveryHour:
		s.Stop()
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{3600, 1}))
		go s.run()
	case EveryDay:
		s.Stop()
		s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, &SyncType{86400, 1}))
		go s.run()
	case DefalutSync:
		s.Stop()
		s.saves = []*SyncType{
			{SecondsSaveSnapshot1, ChangesSaveSnapshot1},
			{SecondsSaveSnapshot2, ChangesSaveSnapshot2},
			{SecondsSaveSnapshot3, ChangesSaveSnapshot3},
		}
		for _, v := range s.saves {
			s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, v))
		}
		go s.run()
	case CustomSync:
		s.Stop()
		for _, v := range s.saves {
			s.snapshotSyncs = append(s.snapshotSyncs, newSnapshotSync(s, v))
		}
		go s.run()
	case Always:
		s.Stop()
		s.always = true
	}
}

func (s *StateMachine) ClearSyncType() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves = []*SyncType{}
}
func (s *StateMachine) AppendSyncType(seconds, changes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves = append(s.saves, &SyncType{seconds, changes})
}
func (s *StateMachine) SetSyncTypes(saves []*SyncType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves = saves
	s.setSnapshotPolicy(CustomSync)
}
func (s *StateMachine) SetSnapshot(snapshot Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snapshot
}

func (s *StateMachine) SaveSnapshot() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveSnapshot()
}
func (s *StateMachine) saveSnapshot() error {
	if s.snapshot != nil {
		if !s.node.storage.Exists(DefaultSnapshot) && s.snapshotReadWriter.work {
			s.snapshotReadWriter.lastIncludedIndex.Set(0)
		}
		if s.lastApplied > s.snapshotReadWriter.lastIncludedIndex.Id() && s.snapshotReadWriter.work {
			Tracef("StateMachine.SaveSnapshot %s Start", s.node.address)
			s.snapshotReadWriter.work = false
			defer func() {
				s.snapshotReadWriter.work = true
			}()
			var lastPrintLastIncludedIndex = s.snapshotReadWriter.lastIncludedIndex.Id()
			var lastIncludedIndex, lastIncludedTerm uint64
			if s.node.lastLogIndex == s.lastApplied {
				lastIncludedIndex = s.node.lastLogIndex
				lastIncludedTerm = s.node.lastLogTerm
			} else {
				lastIncludedIndex = s.lastApplied
				entry := s.node.log.read(lastIncludedIndex)
				if entry == nil {
					return errors.New("this meta is not existed")
				}
				lastIncludedTerm = entry.Term
			}
			s.snapshotReadWriter.Reset(lastIncludedIndex, lastIncludedTerm)
			_, err := s.snapshot.Save(s.node.context, s.snapshotReadWriter)
			s.snapshotReadWriter.Rename()
			startTime := time.Now().UnixNano()
			if s.node.isLeader() {
				s.node.log.deleteBefore(minUint64(s.node.minNextIndex(), lastIncludedIndex))
			} else {
				s.node.log.deleteBefore(lastIncludedIndex)
			}
			go func() {
				if s.node.isLeader() {
					s.snapshotReadWriter.Tar()
				}
			}()
			duration := (time.Now().UnixNano() - startTime) / 1000000
			Tracef("StateMachine.SaveSnapshot %s lastIncludedIndex %d==>%d duration:%dms", s.node.address, lastPrintLastIncludedIndex, s.snapshotReadWriter.lastIncludedIndex.Id(), duration)
			return err
		}
		return nil
	}
	return ErrSnapshotCodecNil
}
func (s *StateMachine) RecoverSnapshot() error {
	if s.snapshot != nil {
		s.snapshotReadWriter.load()
		_, err := s.snapshot.Recover(s.node.context, s.snapshotReadWriter)
		return err
	}
	return ErrSnapshotCodecNil
}

func (s *StateMachine) recover() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.node.storage.IsEmpty(DefaultSnapshot) {
		if !s.node.storage.IsEmpty(DefaultSnapshot + DefaultTmp) {
			s.snapshotReadWriter.Rename()
		} else {
			return errors.New(DefaultSnapshot + " file is empty")
		}
	}
	if s.snapshotReadWriter.lastIncludedIndex.Id() <= s.lastApplied && s.lastApplied > 0 && s.snapshotReadWriter.lastIncludedIndex.Id() > 0 {
		return nil
	}
	err := s.RecoverSnapshot()
	if err != nil {
		s.snapshotReadWriter.lastIncludedIndex.Set(0)
		s.snapshotReadWriter.lastIncludedTerm.Set(0)
		return err
	}
	if s.lastApplied < s.snapshotReadWriter.lastIncludedIndex.Id() {
		var lastApplied = s.lastApplied
		s.lastApplied = s.snapshotReadWriter.lastIncludedIndex.Id()
		Tracef("StateMachine.recover %s lastApplied %d==%d", s.node.address, lastApplied, s.lastApplied)
	}
	if s.node.commitIndex < s.snapshotReadWriter.lastIncludedIndex.Id() {
		var commitIndex = s.node.commitIndex
		s.node.commitIndex = s.snapshotReadWriter.lastIncludedIndex.Id()
		Tracef("StateMachine.recover %s commitIndex %d==%d", s.node.address, commitIndex, s.node.commitIndex)
	}
	return nil
}

func (s *StateMachine) load() {
	s.snapshotReadWriter.load()
}

func (s *StateMachine) append(offset uint64, p []byte) {
	s.snapshotReadWriter.Append(offset, p)
}

func (s *StateMachine) run() {
	Tracef("StateMachine.run %d", len(s.snapshotSyncs))
	for _, snapshotSync := range s.snapshotSyncs {
		go snapshotSync.run()
	}
}
func (s *StateMachine) Stop() {
	Tracef("StateMachine.Stop %d", len(s.snapshotSyncs))
	for _, snapshotSync := range s.snapshotSyncs {
		if snapshotSync != nil {
			snapshotSync.Stop()
			snapshotSync = nil
		}
	}
	s.snapshotSyncs = make([]*snapshotSync, 0)
}
