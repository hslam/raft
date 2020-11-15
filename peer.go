// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"math"
	"sync"
	"sync/atomic"
)

type peer struct {
	mu                 sync.Mutex
	node               *node
	address            string
	alive              bool
	nextIndex          uint64
	lastPrintNextIndex uint64
	checking           int32
	sending            int32
	packing            int32
	installing         int32
	install            int32
	nonVoting          bool
	majorities         bool
	size               uint64
	offset             uint64
	chunk              int
	chunkNum           int
}

func newPeer(n *node, address string) *peer {
	p := &peer{
		node:      n,
		address:   address,
		nextIndex: 0,
	}
	return p
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
	if !p.alive {
		return
	}
	p.alive = p.node.raft.RequestVote(p.address)
}

func (p *peer) appendEntries(entries []*Entry) (nextIndex uint64, term uint64, success bool, ok bool) {
	if !p.alive {
		return
	}
	var prevLogIndex, prevLogTerm uint64
	if p.nextIndex <= 1 {
		prevLogIndex = 0
		prevLogTerm = 0
	} else {
		entry := p.node.log.lookup(p.nextIndex - 1)
		if entry == nil {
			prevLogIndex = 0
			prevLogTerm = 0
		} else {
			prevLogIndex = p.nextIndex - 1
			prevLogTerm = entry.Term
		}
	}
	//logger.Tracef("Peer.run %s %d %d %d ",p.address,prevLogIndex,prevLogTerm,len(entries))
	nextIndex, term, success, ok = p.node.raft.AppendEntries(p.address, prevLogIndex, prevLogTerm, entries)
	if success && ok {
		//logger.Tracef("Peer.run %s nextIndex %d==>%d",p.address,p.nextIndex,nextIndex)
		p.nextIndex = nextIndex
	} else if ok && term == p.node.currentTerm.Load() {
		p.nextIndex = nextIndex
	} else if !ok {
		p.alive = false
	}
	return
}
func (p *peer) installSnapshot(offset uint64, data []byte, Done bool) (recvOffset uint64) {
	if !p.alive {
		return
	}
	var nextIndex uint64
	recvOffset, nextIndex, p.alive = p.node.raft.InstallSnapshot(p.address, p.node.stateMachine.snapshotReadWriter.lastIncludedIndex.ID(), p.node.stateMachine.snapshotReadWriter.lastIncludedTerm.ID(), offset, data, Done)
	if nextIndex > 0 {
		p.nextIndex = nextIndex
	}
	//logger.Debugf("Peer.installSnapshot %s %d %d %t",p.address,offset,len(data),Done)
	return
}

func (p *peer) ping() {
	p.alive = p.node.rpcs.Ping(p.address)
	//logger.Debugf("Peer.ping %s %t",p.address,p.alive)
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
		if ((p.nextIndex == 1 || (p.nextIndex > 1 && p.nextIndex-1 < p.node.firstLogIndex)) && p.node.commitIndex > 1) || p.node.lastLogIndex-(p.nextIndex-1) > defaultNumInstallSnapshot {
			if atomic.CompareAndSwapInt32(&p.installing, 0, 1) {
				atomic.StoreInt32(&p.install, 1)
				//logger.Debugf("Peer.check %s %d %d", p.address, p.nextIndex, p.node.firstLogIndex)
				if p.node.storage.IsEmpty(p.node.stateMachine.snapshotReadWriter.FileName()) {
					if !atomic.CompareAndSwapInt32(&p.packing, 0, 1) {
						atomic.StoreInt32(&p.installing, 0)
						return
					}
					go func() {
						defer atomic.StoreInt32(&p.installing, 0)
						defer atomic.StoreInt32(&p.packing, 0)
						err := p.node.stateMachine.snapshotReadWriter.Tar()
						if err != nil {
							return
						}
					}()
				} else {
					if !atomic.CompareAndSwapInt32(&p.sending, 0, 1) {
						atomic.StoreInt32(&p.installing, 0)
						return
					}
					go func() {
						defer atomic.StoreInt32(&p.installing, 0)
						defer atomic.StoreInt32(&p.sending, 0)
						if p.chunk == 0 {
							size, err := p.node.storage.Size(p.node.stateMachine.snapshotReadWriter.FileName())
							if err != nil {
								return
							}
							p.size = uint64(size)
							p.chunkNum = int(math.Ceil(float64(size) / float64(defaultChunkSize)))
						}
						if p.chunk < p.chunkNum-1 {
							b := make([]byte, defaultChunkSize)
							n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
							if err != nil {
								p.chunk = 0
								p.offset = 0
								atomic.StoreInt32(&p.install, 0)
								return
							}
							if int64(n) == defaultChunkSize {
								offset := p.installSnapshot(p.offset, b[:n], false)
								if offset == p.offset+uint64(n) {
									p.offset += uint64(n)
									p.chunk++
								} else {
									p.chunk = 0
									p.offset = 0
									atomic.StoreInt32(&p.install, 0)
								}
							}
						} else {
							b := make([]byte, p.size-p.offset)
							n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
							if err != nil {
								p.chunk = 0
								p.offset = 0
								atomic.StoreInt32(&p.install, 0)
								return
							}
							if uint64(n) == p.size-p.offset {
								offset := p.installSnapshot(p.offset, b[:n], true)
								if offset == p.offset+uint64(n) {
									p.offset += uint64(n)
									p.chunk++
								} else {
									p.chunk = 0
									p.offset = 0
									atomic.StoreInt32(&p.install, 0)
								}
								//logger.Debugf("Peer.check %s send n-%d offset-%d p.offset-%d", p.address, n, offset, p.offset)
							}
							//logger.Debugf("Peer.check %s send n-%d offset-%d size-%d chunk-%d chunkNum-%d", p.address, n, p.offset, p.size, p.chunk, p.chunkNum)
							if p.offset == p.size && p.chunk == p.chunkNum {
								p.chunk = 0
								p.offset = 0
								atomic.StoreInt32(&p.install, 0)
							}
						}
					}()
				}
			}

		} else {
			if atomic.CompareAndSwapInt32(&p.sending, 0, 1) {
				go func() {
					defer atomic.StoreInt32(&p.sending, 0)
					entries := p.node.log.copyAfter(p.nextIndex, defaultMaxBatch)
					if len(entries) > 0 {
						//logger.Debugf("Peer.check %s send %d %d %d", p.address, p.nextIndex, p.node.firstLogIndex, len(entries))
						p.appendEntries(entries)
						go p.node.commit()
					}
				}()
			}
		}
	}
}
