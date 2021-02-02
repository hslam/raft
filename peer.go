// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/atomic"
	"sync"
	"time"
)

type peer struct {
	mu                 sync.Mutex
	node               *node
	address            string
	alive              *atomic.Bool
	nextIndex          uint64
	matchIndex         uint64
	lastPrintNextIndex uint64
	checking           int32
	sending            int32
	installing         int32
	install            int32
	nonVoting          bool
	majorities         bool
	size               uint64
	offset             uint64
	lastMatchTime      time.Time
}

func newPeer(n *node, address string) *peer {
	p := &peer{
		node:    n,
		address: address,
		alive:   atomic.NewBool(false),
	}
	return p
}

func (p *peer) init() {
	p.nextIndex = p.node.lastLogIndex + 1
	p.matchIndex = 0
	p.install = 0
	p.lastMatchTime = time.Now()
	p.size = 0
	p.offset = 0
}

func (p *peer) heartbeat() bool {
	_, term, success, ok := p.appendEntries([]*Entry{})
	//logger.Tracef("Peer.heartbeat %s %v term-%v success-%v ok-%v ", p.address, p.node.currentTerm.Load(), term, success, ok)
	if ok && success && term == p.node.currentTerm.Load() {
		return true
	}
	return false
}

func (p *peer) requestVote() {
	if !p.alive.Load() {
		return
	}
	ok := p.node.service.raft.CallRequestVote(p.address)
	p.update(ok)
}

func (p *peer) appendEntries(entries []*Entry) (nextIndex uint64, term uint64, success bool, ok bool) {
	if !p.alive.Load() {
		return
	}
	prevLogIndex := p.nextIndex - 1
	prevLogTerm := p.node.log.lookupTerm(prevLogIndex)
	//logger.Tracef("Peer.run %s %d %d %d ",p.address,prevLogIndex,prevLogTerm,len(entries))
	nextIndex, term, success, ok = p.node.service.raft.CallAppendEntries(p.address, prevLogIndex, prevLogTerm, entries)
	if success && ok {
		//logger.Tracef("Peer.run %s nextIndex %d==>%d",p.address,p.nextIndex,nextIndex)
		p.nextIndex = nextIndex
		p.matchIndex = nextIndex - 1
		p.lastMatchTime = time.Now()
	} else if ok && term == p.node.currentTerm.Load() {
		p.nextIndex = nextIndex
		p.matchIndex = 0
		if p.lastMatchTime.Add(defaultMatchTimeout).Before(time.Now()) {
			p.lastMatchTime = time.Now()
			p.matchIndex = 0
			p.nextIndex = 1
		}
	}
	p.update(ok)
	p.checkNextIndex()
	return
}

func (p *peer) installSnapshot(offset uint64, data []byte, Done bool) (recvOffset uint64) {
	if !p.alive.Load() {
		return
	}
	var nextIndex uint64
	var ok bool
	recvOffset, nextIndex, ok = p.node.service.raft.CallInstallSnapshot(p.address, p.node.stateMachine.snapshotReadWriter.lastIncludedIndex.ID(), p.node.stateMachine.snapshotReadWriter.lastIncludedTerm.ID(), offset, data, Done)
	p.update(ok)
	if nextIndex > 0 {
		p.nextIndex = nextIndex
	}
	p.checkNextIndex()
	//logger.Debugf("Peer.installSnapshot %s %d %d %t",p.address,offset,len(data),Done)
	return
}

func (p *peer) ping() {
	err := p.node.rpcs.Ping(p.address)
	p.update(err == nil)
	//logger.Debugf("Peer.ping %s %t",p.address,p.alive)
}

func (p *peer) update(alive bool) {
	if alive {
		if !p.alive.Load() {
			p.alive.Store(true)
		}
	} else {
		if p.alive.Load() {
			p.alive.Store(false)
		}
	}
}

func (p *peer) checkNextIndex() {
	if p.nextIndex > p.node.nextIndex {
		p.nextIndex = p.node.nextIndex
	}
}

func (p *peer) voting() bool {
	return !p.nonVoting && p.majorities
}

func (p *peer) check() {
	if !atomic.CompareAndSwapInt32(&p.checking, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&p.checking, 0)
	if p.node.lastLogIndex > p.nextIndex-1 && p.nextIndex > 0 {
		if p.node.stateMachine.snapshot != nil && p.node.commitIndex.ID() > 1 &&
			(p.nextIndex == 1 || (p.nextIndex > 1 && p.nextIndex < p.node.firstLogIndex)) {
			if !p.node.stateMachine.snapshotReadWriter.finish.Load() {
				return
			}
			if atomic.CompareAndSwapInt32(&p.installing, 0, 1) {
				atomic.AddInt32(&p.install, 1)
				//logger.Debugf("Peer.check %s %d %d", p.address, p.nextIndex, p.node.firstLogIndex)
				go func() {
					defer atomic.StoreInt32(&p.installing, 0)
					if p.size == 0 {
						size, err := p.node.storage.Size(p.node.stateMachine.snapshotReadWriter.FileName())
						if err != nil {
							atomic.StoreInt32(&p.install, 0)
							return
						}
						p.size = uint64(size)
					}
					remain := int64(p.size - p.offset)
					size := defaultChunkSize
					if size > remain {
						size = remain
					}
					b := make([]byte, size)
					n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
					if err == nil && int64(n) == size {
						done := p.size == p.offset+uint64(size)
						offset := p.installSnapshot(p.offset, b, done)
						if offset == p.offset+uint64(n) {
							p.offset += uint64(n)
							if !done {
								return
							}
						}
					}
					p.offset = 0
					p.size = 0
					atomic.StoreInt32(&p.install, 0)
				}()
			}
		} else {
			if p.matchIndex != p.nextIndex-1 {
				return
			}
			if atomic.LoadInt32(&p.install) > 0 {
				atomic.StoreInt32(&p.install, 0)
				p.offset = 0
				p.size = 0
			}
			if atomic.CompareAndSwapInt32(&p.sending, 0, 1) {
				go func() {
					entries := p.node.log.copyAfter(p.nextIndex, defaultMaxBatch)
					if len(entries) > 0 {
						//logger.Debugf("Peer.check %s send %d %d %d", p.address, p.nextIndex, p.node.firstLogIndex, len(entries))
						p.appendEntries(entries)
						go p.node.commit()
					}
					atomic.StoreInt32(&p.sending, 0)
				}()
			}
		}
	}
}
