// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"github.com/hslam/wal"
	"sync"
	"time"
)

//
type waLog struct {
	mu               sync.Mutex
	batchMu          sync.Mutex
	pauseMu          sync.Mutex
	node             *node
	wal              *wal.Log
	compactionTicker *time.Ticker
	buf              []byte
	entryPool        *sync.Pool
	stop             chan bool
	finish           chan bool
	work             bool
	paused           bool
}

func newLog(n *node) *waLog {
	l := &waLog{
		node:             n,
		compactionTicker: time.NewTicker(defaultCompactionTick),
		buf:              make([]byte, 1024*64),
		stop:             make(chan bool, 1),
		finish:           make(chan bool, 1),
		work:             true,
	}
	l.wal, _ = wal.Open(n.storage.dataDir, nil)
	l.entryPool = &sync.Pool{
		New: func() interface{} {
			return &Entry{}
		},
	}
	go l.run()
	return l
}
func (l *waLog) getEmtyEntry() *Entry {
	return l.entryPool.Get().(*Entry)
}
func (l *waLog) putEmtyEntry(entry *Entry) {
	entry.Index = 0
	entry.Term = 0
	entry.Command = []byte{}
	entry.CommandType = 0
	l.entryPool.Put(entry)
}
func (l *waLog) putEmtyEntries(entries []*Entry) {
	for _, entry := range entries {
		l.putEmtyEntry(entry)
	}
}

func (l *waLog) pause(p bool) {
	l.pauseMu.Lock()
	defer l.pauseMu.Unlock()
	l.paused = p
}
func (l *waLog) isPaused() bool {
	l.pauseMu.Lock()
	defer l.pauseMu.Unlock()
	return l.paused
}
func (l *waLog) checkPaused() {
	for {
		if !l.isPaused() {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}
func (l *waLog) checkIndex(index uint64) bool {
	ok, err := l.wal.IsExist(index)
	if err != nil {
		return false
	}
	return ok
}

func (l *waLog) lookup(index uint64) *Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	return l.read(index)
}

func (l *waLog) consistencyCheck(index uint64, term uint64) (ok bool) {
	if index == 0 {
		return false
	}
	if index > l.node.lastLogIndex {
		return false
	}
	if index == l.node.lastLogIndex {
		if term == l.node.lastLogTerm {
			return true
		}
	}
	entry := l.node.log.lookup(index)
	if entry == nil {
		return false
	}
	if entry.Term != term {
		l.node.log.deleteAfter(index)
		l.node.nextIndex = l.node.lastLogIndex + 1
		return false
	}
	return true
}
func (l *waLog) check(entries []*Entry) bool {
	lastIndex := entries[0].Index
	for i := 1; i < len(entries); i++ {
		if entries[i].Index == lastIndex+1 {
			lastIndex = entries[i].Index
		} else {
			return false
		}
	}
	return true
}

func (l *waLog) deleteAfter(index uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pause(true)
	defer l.pause(false)
	if index == l.node.firstLogIndex {
		l.wal.Reset()
		l.wal.InitFirstIndex(index)
		return
	}
	l.node.lastLogIndex = index - 1
	l.wal.Truncate(index - 1)
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex > 0 {
		entry := l.read(lastLogIndex)
		l.node.lastLogTerm = entry.Term
	} else {
		l.node.lastLogTerm = 0
	}
}

func (l *waLog) deleteBefore(index uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	l.pause(true)
	defer l.pause(false)
	l.wal.Clean(index)
	l.node.firstLogIndex, _ = l.wal.FirstIndex()
	logger.Tracef("l.deleteBefore %s deleteBefore %d", l.node.address, index)
}
func (l *waLog) startIndex(index uint64) uint64 {
	return maxUint64(l.node.firstLogIndex, index)
}
func (l *waLog) endIndex(index uint64) uint64 {
	return minUint64(index, l.node.lastLogIndex)
}
func (l *waLog) copyAfter(index uint64, max int) (entries []*Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	startIndex := l.startIndex(index)
	endIndex := l.endIndex(startIndex + uint64(max))
	return l.copyRange(startIndex, endIndex)
}
func (l *waLog) copyRange(startIndex uint64, endIndex uint64) []*Entry {
	//logger.Tracef("l.copyRange %s startIndex %d endIndex %d",l.node.address,metas[0].Index,metas[len(metas)-1].Index)
	return l.batchRead(startIndex, endIndex)
}

func (l *waLog) applyCommited() {
	l.node.stateMachine.Lock()
	defer l.node.stateMachine.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(l.node.stateMachine.lastApplied+1, 1)
	var endIndex = l.node.commitIndex
	if startIndex > endIndex {
		return
	}
	if endIndex-startIndex > defaultMaxBatch {
		index := startIndex
		for {
			l.applyCommitedRange(index, index+defaultMaxBatch)
			index += defaultMaxBatch
			if endIndex-index <= defaultMaxBatch {
				l.applyCommitedRange(index, endIndex)
				break
			}
		}
	} else {
		l.applyCommitedRange(startIndex, endIndex)
	}
}

func (l *waLog) applyCommitedEnd(endIndex uint64) {
	l.node.stateMachine.Lock()
	defer l.node.stateMachine.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(l.node.stateMachine.lastApplied+1, 1)
	endIndex = minUint64(l.node.commitIndex, endIndex)
	if startIndex > endIndex {
		return
	}
	if endIndex-startIndex > defaultMaxBatch {
		index := startIndex
		for {
			l.applyCommitedRange(index, index+defaultMaxBatch)
			index += defaultMaxBatch
			if endIndex-index <= defaultMaxBatch {
				l.applyCommitedRange(index, endIndex)
				break
			}
		}
	} else {
		l.applyCommitedRange(startIndex, endIndex)
	}
}

func (l *waLog) applyCommitedRange(startIndex uint64, endIndex uint64) {
	//logger.Tracef("l.applyCommitedRange %s startIndex %d endIndex %d Start",l.node.address,startIndex,endIndex)
	entries := l.copyRange(startIndex, endIndex)
	if entries == nil || len(entries) == 0 {
		return
	}
	//logger.Tracef("l.applyCommitedRange %s startIndex %d endIndex %d length %d",l.node.address,startIndex,endIndex,len(entries))
	for i := 0; i < len(entries); i++ {
		//logger.Tracef("l.applyCommitedRange %s Index %d Type %d",l.node.address,entries[i].Index,entries[i].CommandType)
		command := l.node.commands.clone(entries[i].CommandType)
		var err error
		if entries[i].CommandType >= 0 {
			err = l.node.codec.Unmarshal(entries[i].Command, command)
		} else {
			err = l.node.raftCodec.Unmarshal(entries[i].Command, command)
		}
		if err == nil {
			l.node.stateMachine.apply(entries[i].Index, command)
		} else {
			logger.Errorf("l.applyCommitedRange %s %d error %s", l.node.address, i, err)
		}
		l.node.commands.put(command)
	}
	l.putEmtyEntries(entries)
	//logger.Tracef("l.applyCommitedRange %s startIndex %d endIndex %d End %d",l.node.address,startIndex,endIndex,len(entries))
}

func (l *waLog) appendEntries(entries []*Entry) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	if !l.node.isLeader() {
		if !l.check(entries) {
			return false
		}
	}
	if l.node.lastLogIndex != entries[0].Index-1 {
		return false
	}
	l.Write(entries)
	l.node.lastLogIndex = entries[len(entries)-1].Index
	l.node.lastLogTerm = entries[len(entries)-1].Term
	l.putEmtyEntries(entries)
	//logger.Tracef("l.appendEntries %s entries %d", l.node.address, len(entries))
	return true
}
func (l *waLog) read(index uint64) *Entry {
	b, err := l.wal.Read(index)
	if err != nil {
		return nil
	}
	entry := l.getEmtyEntry()
	err = l.node.raftCodec.Unmarshal(b, entry)
	if err != nil {
		logger.Errorf("l.Decode %s", string(b))
		return nil
	}
	return entry
}

func (l *waLog) batchRead(startIndex uint64, endIndex uint64) []*Entry {
	entries := make([]*Entry, 0, endIndex-startIndex+1)
	for i := startIndex; i < endIndex+1; i++ {
		entries = append(entries, l.read(i))
	}
	return entries
}

func (l *waLog) load() (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.checkPaused()
	lastLogIndex := l.node.lastLogIndex
	l.node.firstLogIndex, err = l.wal.FirstIndex()
	if err != nil {
		return err
	}
	l.node.lastLogIndex, err = l.wal.LastIndex()
	if err != nil {
		return err
	}
	if l.node.lastLogIndex > 0 {
		entry := l.read(l.node.lastLogIndex)
		if entry != nil {
			l.node.lastLogTerm = entry.Term
			l.node.recoverLogIndex = l.node.lastLogIndex
			l.node.nextIndex = l.node.lastLogIndex + 1
		}
	}
	if l.node.lastLogIndex > lastLogIndex {
		logger.Tracef("l.recover %s lastLogIndex %d", l.node.address, lastLogIndex)
	}
	return nil
}

func (l *waLog) Write(entries []*Entry) (err error) {
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		b, err := l.node.raftCodec.Marshal(l.buf, entry)
		if err != nil {
			return err
		}
		if err = l.wal.Write(entry.Index, b); err != nil {
			return err
		}
	}
	//logger.Tracef("l.Write %d", len(entries))
	return l.wal.FlushAndSync()
}

func (l *waLog) compaction() error {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	//md5, err:=l.loadMd5()
	//if err==nil&&md5!=""{
	//	if md5==l.node.storage.MD5(DefaultLog){
	//		return nil
	//	}
	//}
	//var compactionEntries =make(map[string]*Entry)
	//var entries	=make([]*Entry,0)
	//for i:=len(l.entries)-1;i>=0;i--{
	//	entry:= l.entries[i]
	//	key:=string(int32ToBytes(entry.CommandType))+entry.CommandId
	//	if _,ok:=compactionEntries[key];!ok{
	//		entries=append(entries,entry)
	//		compactionEntries[key]=entry
	//	}
	//}
	//for i,j:= 0,len(entries)-1;i<j;i,j=i+1,j-1{
	//	entries[i], entries[j] = entries[j], entries[i]
	//}
	//l.entries=entries
	//
	//b,metas,_:=l.Encode(l.entries,0)
	//var lastLogSize ,_= l.node.storage.Size(DefaultLog)
	//l.node.storage.SafeOverWrite(DefaultLog,b)
	//l.indexs.metas=metas
	//l.indexs.save()
	//l.ret=uint64(len(b))
	//if len(l.entries)>0{
	//	l.node.lastLogIndex=l.entries[len(l.entries)-1].Index
	//	l.node.lastLogTerm=l.entries[len(l.entries)-1].Term
	//}
	//var logSize ,_= l.node.storage.Size(DefaultLog)
	//logger.Tracef("l.compaction %s LogSize %d==>%d",l.node.address,lastLogSize,logSize)
	//l.saveMd5()
	//logger.Tracef("l.compaction %s md5 %s==>%s",l.node.address,md5,l.node.storage.MD5(DefaultLog))
	return nil
}
func (l *waLog) saveMd5() {
	//md5 := l.node.storage.MD5(DefaultLog)
	//if len(md5) > 0 {
	//	l.node.storage.OverWrite(DefaultMd5, []byte(md5))
	//}
}

func (l *waLog) loadMd5() (string, error) {
	if !l.node.storage.Exists(defaultMd5) {
		return "", errors.New("md5 file is not existed")
	}
	b, err := l.node.storage.Load(defaultMd5)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (l *waLog) run() {
	for {
		select {
		case <-l.compactionTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				//l.compaction()
			}()
		case <-l.stop:
			close(l.stop)
			l.stop = nil
			goto endfor
		}
	}
endfor:
	l.compactionTicker.Stop()
	l.compactionTicker = nil
	l.finish <- true
}
func (l *waLog) ticker(entries []*Entry) {
	l.appendEntries(entries)
}
func (l *waLog) Stop() {
	if l.stop == nil {
		return
	}
	l.stop <- true
	select {
	case <-l.finish:
		close(l.finish)
		l.finish = nil
	}
}
