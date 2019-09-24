package raft

import (
	"sync"
	"time"
	"errors"
)

type StateMachine struct {
	mu 								sync.RWMutex
	node 							*Node
	lastApplied						uint64
	lastSaved						uint64
	snapshot						Snapshot
	snapshotSyncType				SnapshotSyncType
	snapshotSyncTicker				*time.Ticker
}
func newStateMachine(node *Node)*StateMachine {
	s:=&StateMachine{
		node:node,
		snapshotSyncType:Never,
		snapshotSyncTicker:time.NewTicker(time.Second),
	}
	s.load()
	go s.run()
	return s
}
func (s *StateMachine)Apply(index uint64,command Command) (interface{},error){
	s.mu.Lock()
	defer s.mu.Unlock()
	if index<=s.lastApplied{
		return nil,nil
	}
	reply,err:=command.Do(s.node.context)
	s.lastApplied=index
	return reply,err
}

func (s *StateMachine)SetSnapshotSyncType(snapshotSyncType SnapshotSyncType){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotSyncType=snapshotSyncType
}
func (s *StateMachine)SetSnapshot(snapshot Snapshot){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot=snapshot
}


func (s *StateMachine)InstallSnapshot(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot!=nil{
		s.save(data)
		s.RecoverSnapshot(data)
	}
	return ErrSnapshotCodecNil
}
func (s *StateMachine)SaveSnapshot()error{
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot!=nil{
		b,err:=s.snapshot.Save(s.node.context)
		if err==nil{
			return s.save(b)
		}
		return err
	}
	return ErrSnapshotCodecNil
}
func (s *StateMachine)RecoverSnapshot(data []byte) error {
	if s.snapshot!=nil{
		return s.snapshot.Recover(s.node.context,data)
	}
	return ErrSnapshotCodecNil
}

func (s *StateMachine)save(data []byte) error {
	if s.snapshot!=nil{
		s.saveLastSaved()
		return s.node.storage.SafeOverWrite(DefaultSnapshot,data)
	}
	return ErrSnapshotCodecNil
}
func (s *StateMachine) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.node.storage.Exists(DefaultSnapshot){
		return nil
	}
	b, err := s.node.storage.Load(DefaultSnapshot)
	if err != nil {
		return err
	}
	s.loadLastSaved()
	s.RecoverSnapshot(b)
	return nil
}

func (s *StateMachine) saveLastSaved() {
	s.lastSaved=s.lastApplied
	s.node.storage.SafeOverWrite(DefaultLastSaved,uint64ToBytes(s.lastSaved))
}

func (s *StateMachine) loadLastSaved() error {
	if !s.node.storage.Exists(DefaultLastSaved){
		return errors.New("lastsaved file is not existed")
	}
	b, err := s.node.storage.Load(DefaultLastSaved)
	if err != nil {
		return err
	}
	s.lastSaved= bytesToUint64(b)
	if s.node.commitIndex.Id()<s.lastSaved{
		var commitIndex=s.lastApplied
		s.node.commitIndex.Set(s.lastSaved)
		Tracef("StateMachine.loadLastSaved %s recover commitIndex %d==%d",s.node.address,commitIndex,s.node.commitIndex.Id())
	}
	if s.lastApplied<s.lastSaved{
		var lastApplied=s.lastApplied
		s.lastApplied=s.lastSaved
		Tracef("StateMachine.loadLastSaved %s recover lastApplied %d==%d",s.node.address,lastApplied,s.lastApplied)
	}
	return nil
}
func (s *StateMachine)run()  {
	for{
		select {
		case <-s.snapshotSyncTicker.C:
			if s.snapshotSyncType==EverySecond{
				s.SaveSnapshot()
			}
		}
	}
}