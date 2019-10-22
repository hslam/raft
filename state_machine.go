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
	snapshot						Snapshot
	snapshotReadWriter				*SnapshotReadWriter
	snapshotSyncType				SnapshotSyncType
	snapshotSyncTicker				*time.Ticker
	stop 							chan bool
	finish 							chan bool
	saves 							[][]int
	saveLog 						bool
}
func newStateMachine(node *Node)*StateMachine {
	s:=&StateMachine{
		node:node,
		snapshotReadWriter:newSnapshotReadWriter(node,DefaultSnapshot),
		saves:[][]int{
			{SecondsSaveSnapshot1,ChangesSaveSnapshot1},
			{SecondsSaveSnapshot2,ChangesSaveSnapshot2},
			{SecondsSaveSnapshot3,ChangesSaveSnapshot3},
			},
	}
	s.SetSnapshotSyncType(Never)
	return s
}
func (s *StateMachine)Apply(index uint64,command Command) (interface{},error){
	s.mu.Lock()
	defer s.mu.Unlock()
	if index<=s.lastApplied{
		return nil,nil
	}
	var reply interface{}
	var err error
	if command.Type()>=0{
		reply,err=command.Do(s.node.context)
	}else {
		reply,err=command.Do(s.node)
	}
	s.lastApplied=index
	return reply,err
}

func (s *StateMachine)SetSnapshotSyncType(snapshotSyncType SnapshotSyncType){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotSyncType=snapshotSyncType
	switch s.snapshotSyncType {
	case Never:
	case EverySecond:
		s.Stop()
		s.snapshotSyncTicker=time.NewTicker(time.Minute)
		s.stop=make(chan bool,1)
		s.finish=make(chan bool,1)
		go s.run()
	case EveryMinute:
		s.Stop()
		s.snapshotSyncTicker=time.NewTicker(time.Minute)
		s.stop=make(chan bool,1)
		s.finish=make(chan bool,1)
		go s.run()
	case EveryHour:
		s.Stop()
		s.snapshotSyncTicker=time.NewTicker(time.Hour)
		s.stop=make(chan bool,1)
		s.finish=make(chan bool,1)
		go s.run()
	case EveryDay:
		s.Stop()
		s.snapshotSyncTicker=time.NewTicker(time.Hour*24)
		s.stop=make(chan bool,1)
		s.finish=make(chan bool,1)
		go s.run()
	case DefalutSave:
	case Save:
	case Always:
	}
}

func (s *StateMachine)AppendSave(seconds,changes int){
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *StateMachine)SetSnapshot(snapshot Snapshot){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot=snapshot
}


func (s *StateMachine)InstallSnapshot(p []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_,err:=s.snapshotReadWriter.Write(p)
	return err
}
func (s *StateMachine)SaveSnapshot()error{
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot!=nil{
		if !s.node.storage.Exists(DefaultSnapshot)&&s.snapshotReadWriter.work{
			s.snapshotReadWriter.lastIncludedIndex.Set(0)
		}
		if s.lastApplied>s.snapshotReadWriter.lastIncludedIndex.Id()&&s.snapshotReadWriter.work{
			s.snapshotReadWriter.work=false
			defer func() {
				s.snapshotReadWriter.work=true
			}()
			var lastPrintLastIncludedIndex =s.snapshotReadWriter.lastIncludedIndex.Id()
			var lastIncludedIndex,lastIncludedTerm uint64
			if s.node.lastLogIndex==s.lastApplied{
				lastIncludedIndex=s.node.lastLogIndex
				lastIncludedTerm=s.node.lastLogTerm
			}else {
				lastIncludedIndex=s.lastApplied
				meta:=s.node.log.indexs.lookup(lastIncludedIndex)
				if meta==nil{
					return errors.New("this meta is not existed")
				}
				lastIncludedTerm=meta.Term
			}
			s.snapshotReadWriter.Reset(lastIncludedIndex,lastIncludedTerm)
			_,err:=s.snapshot.Save(s.node.context,s.snapshotReadWriter)
			Tracef("StateMachine.SaveSnapshot %s lastIncludedIndex %d==%d",s.node.address,lastPrintLastIncludedIndex,s.snapshotReadWriter.lastIncludedIndex.Id())
			return err
		}
		return nil
	}
	return ErrSnapshotCodecNil
}
func (s *StateMachine)RecoverSnapshot() error {
	if s.snapshot!=nil{
		s.snapshotReadWriter.load()
		_,err:=s.snapshot.Recover(s.node.context,s.snapshotReadWriter)
		return err
	}
	return ErrSnapshotCodecNil
}

func (s *StateMachine) recover() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.node.storage.IsEmpty(DefaultSnapshot){
		return errors.New(DefaultSnapshot+" file is empty")
	}
	if s.snapshotReadWriter.lastIncludedIndex.Id()<=s.lastApplied&&s.lastApplied>0&&s.snapshotReadWriter.lastIncludedIndex.Id()>0{
		return nil
	}
	err:=s.RecoverSnapshot()
	if err!=nil{
		s.snapshotReadWriter.lastIncludedIndex.Set(0)
		s.snapshotReadWriter.lastIncludedTerm.Set(0)
		return err
	}
	if s.lastApplied<s.snapshotReadWriter.lastIncludedIndex.Id(){
		var lastApplied=s.lastApplied
		s.lastApplied=s.snapshotReadWriter.lastIncludedIndex.Id()
		Tracef("StateMachine.recover %s lastApplied %d==%d",s.node.address,lastApplied,s.lastApplied)
	}
	if s.node.commitIndex.Id()<s.snapshotReadWriter.lastIncludedIndex.Id(){
		var commitIndex=s.node.commitIndex.Id()
		s.node.commitIndex.Set(s.snapshotReadWriter.lastIncludedIndex.Id())
		Tracef("StateMachine.recover %s commitIndex %d==%d",s.node.address,commitIndex,s.node.commitIndex.Id())
	}
	return nil
}

func (s *StateMachine) load() {
	s.snapshotReadWriter.load()
}

func (s *StateMachine) append(offset uint64,p []byte) {
	s.snapshotReadWriter.Append(offset,p)
}

func (s *StateMachine)run()  {
	go func() {
		for range s.snapshotSyncTicker.C{
			if s.snapshotSyncType==EverySecond||s.snapshotSyncType==EveryMinute||s.snapshotSyncType==EveryHour{
				s.SaveSnapshot()
			}
		}
	}()
	select {
	case <-s.stop:
		close(s.stop)
		s.stop=nil
	}
	s.snapshotSyncTicker.Stop()
	s.snapshotSyncTicker=nil
	s.finish<-true
}
func (s *StateMachine)Stop()  {
	if s.stop==nil{
		return
	}
	s.stop<-true
	select {
	case <-s.finish:
		close(s.finish)
		s.finish=nil
	}
}