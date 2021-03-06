// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/lru"
	"github.com/hslam/wal"
	"sync"
)

// WAL represents the write-ahead log.
type WAL interface {
	FirstIndex() (uint64, error)
	LastIndex() (uint64, error)
	Read(index uint64) ([]byte, error)
	Write(index uint64, data []byte) error
	Flush() error
	Sync() error
	Clean(index uint64) error
	Truncate(index uint64) error
	Reset() error
	Close() error
}

type waLog struct {
	mu     sync.Mutex
	node   *node
	wal    WAL
	lru    *lru.LRU
	cache  *cache
	buf    []byte
	thresh uint64
}

func newLog(n *node) *waLog {
	l := &waLog{
		node:   n,
		buf:    make([]byte, 1024*64),
		lru:    lru.New(defaultMaxCache, nil),
		cache:  newCache(defaultMaxCache),
		thresh: 1,
	}
	l.wal, _ = wal.Open(n.storage.dataDir, nil)
	return l
}

func (l *waLog) cloneEntry(entry *Entry) *Entry {
	clone := &Entry{}
	*clone = *entry
	clone.Command = cloneBytes(entry.Command)
	return clone
}

func (l *waLog) lookupTerm(index uint64) (term uint64) {
	l.mu.Lock()
	term = l.readTerm(index)
	l.mu.Unlock()
	return
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
	entryTerm := l.node.log.lookupTerm(index)
	if entryTerm == 0 {
		return false
	}
	if entryTerm != term {
		if index-l.thresh >= l.node.firstLogIndex {
			index -= l.thresh
			l.thresh *= 2
		} else {
			index = l.node.firstLogIndex
		}
		l.node.log.deleteAfter(index)
		l.node.nextIndex = l.node.lastLogIndex + 1
		return false
	}
	l.thresh = 1
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
	l.lru.Reset()
	if index == l.node.firstLogIndex {
		l.wal.Reset()
		l.mu.Unlock()
		return
	}
	l.node.lastLogIndex = index - 1
	l.wal.Truncate(index - 1)
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex > 0 {
		entryTerm := l.readTerm(lastLogIndex)
		l.node.lastLogTerm = entryTerm
	} else {
		l.node.lastLogTerm = 0
	}
	l.mu.Unlock()
}

func (l *waLog) deleteBefore(index uint64) {
	l.mu.Lock()
	l.wal.Clean(index)
	l.node.firstLogIndex, _ = l.wal.FirstIndex()
	l.mu.Unlock()
	l.node.logger.Tracef("log.clean %s index %d", l.node.address, index)
}

func (l *waLog) startIndex(index uint64) uint64 {
	return maxUint64(l.node.firstLogIndex, index)
}

func (l *waLog) endIndex(index uint64) uint64 {
	return minUint64(index, l.node.lastLogIndex)
}

func (l *waLog) copyAfter(index uint64, max int) (entries []*Entry) {
	entries = l.cache.CopyAfter(index, max)
	if len(entries) > 0 {
		return
	}
	l.mu.Lock()
	startIndex := l.startIndex(index)
	endIndex := l.endIndex(startIndex + uint64(max))
	if startIndex <= endIndex {
		entries = l.copyRange(startIndex, endIndex)
	}
	l.mu.Unlock()
	return
}

func (l *waLog) copyRange(startIndex uint64, endIndex uint64) []*Entry {
	//l.node.logger.Tracef("log.copyRange %s startIndex %d endIndex %d",l.node.address,metas[0].Index,metas[len(metas)-1].Index)
	return l.batchRead(startIndex, endIndex)
}

func (l *waLog) applyCommited() {
	l.node.stateMachine.Lock()
	defer l.node.stateMachine.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(l.node.stateMachine.lastApplied+1, l.node.firstLogIndex)
	var endIndex = minUint64(l.node.commitIndex.ID(), l.node.lastLogIndex)
	l.applyCommitedRange(startIndex, endIndex)
}

func (l *waLog) applyCommitedEnd(endIndex uint64) {
	l.node.stateMachine.Lock()
	defer l.node.stateMachine.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	lastLogIndex := l.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(l.node.stateMachine.lastApplied+1, l.node.firstLogIndex)
	endIndex = minUint64(minUint64(l.node.commitIndex.ID(), l.node.lastLogIndex), endIndex)
	l.applyCommitedRange(startIndex, endIndex)
}

func (l *waLog) applyCommitedRange(startIndex uint64, endIndex uint64) {
	if startIndex > endIndex {
		return
	}
	if endIndex-startIndex > defaultMaxBatch {
		index := startIndex
		for {
			l.applyCommitedBatch(index, index+defaultMaxBatch)
			index += defaultMaxBatch
			if endIndex-index <= defaultMaxBatch {
				l.applyCommitedBatch(index, endIndex)
				break
			}
		}
	} else {
		l.applyCommitedBatch(startIndex, endIndex)
	}
}

func (l *waLog) applyCommitedBatch(startIndex uint64, endIndex uint64) {
	//l.node.logger.Tracef("log.applyCommitedRange %s startIndex %d endIndex %d Start", l.node.address, startIndex, endIndex)
	entries := l.copyRange(startIndex, endIndex)
	if entries == nil || len(entries) == 0 {
		return
	}
	//l.node.logger.Tracef("log.applyCommitedRange %s startIndex %d endIndex %d length %d",l.node.address,startIndex,endIndex,len(entries))
	for i := 0; i < len(entries); i++ {
		//l.node.logger.Tracef("log.applyCommitedRange %s Index %d Type %d",l.node.address,entries[i].Index,entries[i].CommandType)
		entry := entries[i]
		command := l.node.commands.clone(entry.CommandType)
		var err error
		if entry.CommandType > 0 {
			err = l.node.codec.Unmarshal(entry.Command, command)
		} else {
			err = l.node.raftCodec.Unmarshal(entry.Command, command)
		}
		if err == nil {
			l.node.stateMachine.apply(entry.Index, command)
		} else {
			l.node.logger.Errorf("log.applyCommitedRange %s %d error %s", l.node.address, i, err.Error())
		}
		l.node.commands.put(command)
	}
	//l.node.logger.Tracef("log.applyCommitedRange %s startIndex %d endIndex %d End %d",l.node.address,startIndex,endIndex,len(entries))
}

func (l *waLog) readTerm(index uint64) (term uint64) {
	entry := l.read(index)
	if entry != nil {
		term = entry.Term
	} else if l.node.stateMachine.snapshotReadWriter.lastIncludedIndex.ID() == index {
		term = l.node.stateMachine.snapshotReadWriter.lastIncludedTerm.ID()
	}
	return
}

func (l *waLog) Read(index uint64) *Entry {
	l.mu.Lock()
	entry := l.read(index)
	l.mu.Unlock()
	return entry
}

func (l *waLog) read(index uint64) *Entry {
	if value, _, ok := l.lru.Get(index); ok {
		if entry, ok := value.(*Entry); ok && entry.Index == index {
			return entry
		}
	}
	b, err := l.wal.Read(index)
	if err != nil {
		//l.node.logger.Errorf("log.read %s index %d firstIndex %d lastIndex %d, commit %d", l.node.address, index, l.node.firstLogIndex, l.node.lastLogIndex, l.node.commitIndex.ID())
		return nil
	}
	entry := &Entry{}
	err = l.node.codec.Unmarshal(b, entry)
	if err != nil {
		l.node.logger.Errorf("log.Decode %s", string(b))
		return nil
	}
	l.lru.Set(entry.Index, entry)
	return entry
}

func (l *waLog) batchRead(startIndex uint64, endIndex uint64) []*Entry {
	entries := make([]*Entry, 0, endIndex-startIndex+1)
	for i := startIndex; i < endIndex+1; i++ {
		entry := l.read(i)
		if entry == nil {
			return nil

		}
		entries = append(entries, entry)
	}
	return entries
}

func (l *waLog) load() (err error) {
	l.mu.Lock()
	lastLogIndex := l.node.lastLogIndex
	l.node.firstLogIndex, err = l.wal.FirstIndex()
	if err != nil {
		l.mu.Unlock()
		return err
	}
	l.node.lastLogIndex, err = l.wal.LastIndex()
	if err != nil {
		l.mu.Unlock()
		return err
	}
	if l.node.lastLogIndex > 0 {
		entry := l.read(l.node.lastLogIndex)
		if entry != nil {
			l.node.lastLogTerm = entry.Term
			l.node.nextIndex = l.node.lastLogIndex + 1
		}
	}
	if l.node.lastLogIndex > lastLogIndex {
		l.node.logger.Tracef("log.recover %s lastLogIndex %d", l.node.address, lastLogIndex)
	}
	l.mu.Unlock()
	return nil
}

func (l *waLog) appendEntries(entries []*Entry, clone bool) bool {
	l.mu.Lock()
	if !l.node.isLeader() {
		if !l.check(entries) {
			l.mu.Unlock()
			return false
		}
	}
	if l.node.lastLogIndex != entries[0].Index-1 {
		l.mu.Unlock()
		return false
	}
	l.Write(entries, clone)
	l.node.lastLogIndex = entries[len(entries)-1].Index
	l.node.lastLogTerm = entries[len(entries)-1].Term
	l.mu.Unlock()
	//l.node.logger.Tracef("log.appendEntries %s entries %d", l.node.address, len(entries))
	return true
}

func (l *waLog) Write(entries []*Entry, clone bool) (err error) {
	for i := 0; i < len(entries); i++ {
		var entry *Entry
		if clone {
			entry = l.cloneEntry(entries[i])
		} else {
			entry = entries[i]
		}
		l.lru.Set(entry.Index, entry)
		b, err := l.node.codec.Marshal(l.buf, entry)
		if err != nil {
			return err
		}
		if err = l.wal.Write(entry.Index, b); err != nil {
			return err
		}
	}
	//l.node.logger.Tracef("log.Write %d", len(entries))
	err = l.wal.Flush()
	if err != nil {
		return
	}
	return l.wal.Sync()
}

func (l *waLog) Stop() (err error) {
	l.cache.Reset()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lru.Reset()
	return l.wal.Close()
}
