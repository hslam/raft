package raft

import (
	"math"
	"sync"
)

type Peer struct {
	mu                 sync.Mutex
	node               *Node
	address            string
	alive              bool
	nextIndex          uint64
	lastPrintNextIndex uint64
	send               bool
	tarWork            bool
	install            bool
	nonVoting          bool
	majorities         bool
	size               uint64
	offset             uint64
	chunk              int
	chunkNum           int
}

func newPeer(node *Node, address string) *Peer {
	p := &Peer{
		node:      node,
		address:   address,
		nextIndex: 0,
		send:      true,
		tarWork:   true,
		install:   true,
	}
	return p
}

func (p *Peer) heartbeat() {
	p.appendEntries([]*Entry{})
}

func (p *Peer) requestVote() {
	if !p.alive {
		return
	}
	p.alive = p.node.raft.RequestVote(p.address)
}

func (p *Peer) appendEntries(entries []*Entry) (nextIndex uint64, term uint64, success bool, ok bool) {
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
	//Tracef("Peer.run %s %d %d %d ",p.address,prevLogIndex,prevLogTerm,len(entries))
	nextIndex, term, success, ok = p.node.raft.AppendEntries(p.address, prevLogIndex, prevLogTerm, entries)
	if success && ok {
		//Tracef("Peer.run %s nextIndex %d==>%d",p.address,p.nextIndex,nextIndex)
		p.nextIndex = nextIndex
	} else if ok && term == p.node.currentTerm.Id() {
		p.nextIndex = nextIndex
	} else if !ok {
		p.alive = false
	}
	return
}
func (p *Peer) installSnapshot(offset uint64, data []byte, Done bool) (recv_offset uint64) {
	if !p.alive {
		return
	}
	var nextIndex uint64
	recv_offset, nextIndex, p.alive = p.node.raft.InstallSnapshot(p.address, p.node.stateMachine.snapshotReadWriter.lastIncludedIndex.Id(), p.node.stateMachine.snapshotReadWriter.lastIncludedTerm.Id(), offset, data, Done)
	if nextIndex > 0 {
		p.nextIndex = nextIndex
	}
	//Debugf("Peer.installSnapshot %s %d %d %t",p.address,offset,len(data),Done)
	return
}

func (p *Peer) ping() {
	p.alive = p.node.rpcs.Ping(p.address)
	//Debugf("Peer.ping %s %t",p.address,p.alive)
}

func (p *Peer) voting() bool {
	return !p.nonVoting && p.majorities
}

func (p *Peer) check() {
	if p.node.lastLogIndex > p.nextIndex-1 && p.nextIndex > 0 {
		if ((p.nextIndex == 1 || (p.nextIndex > 1 && p.nextIndex-1 < p.node.firstLogIndex)) && p.node.commitIndex > 1) || p.node.lastLogIndex-(p.nextIndex-1) > DefaultNumInstallSnapshot {
			if p.install {
				p.install = false
				defer func() {
					p.install = true
				}()
				if p.node.storage.IsEmpty(p.node.stateMachine.snapshotReadWriter.FileName()) && p.tarWork {
					p.tarWork = false
					go func() {
						defer func() {
							p.tarWork = true
						}()
						err := p.node.stateMachine.snapshotReadWriter.Tar()
						if err != nil {
							return
						}
					}()
				} else {
					if p.send {
						p.send = false
						go func() {
							defer func() {
								p.send = true
							}()
							if p.chunk == 0 {
								size, err := p.node.storage.Size(p.node.stateMachine.snapshotReadWriter.FileName())
								if err != nil {
									return
								}
								p.size = uint64(size)
								p.chunkNum = int(math.Ceil(float64(size) / float64(DefaultChunkSize)))
							}
							if p.chunkNum > 1 {
								if p.chunk < p.chunkNum-1 {
									b := make([]byte, DefaultChunkSize)
									n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
									if err != nil {
										return
									}
									if int64(n) == DefaultChunkSize {
										offset := p.installSnapshot(p.offset, b[:n], false)
										if offset == p.offset+uint64(n) {
											p.offset += uint64(n)
											p.chunk += 1
										}
									}
								} else {
									b := make([]byte, p.size-p.offset)
									n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
									if err != nil {
										return
									}
									if uint64(n) == p.size-p.offset {
										offset := p.installSnapshot(p.offset, b[:n], true)
										if offset == p.offset+uint64(n) {
											p.offset += uint64(n)
											p.chunk += 1
										}
									}
									if p.offset == p.size && p.chunk == p.chunkNum {
										p.chunk = 0
										p.offset = 0
									}
								}
							} else {
								b := make([]byte, p.size)
								p.offset = 0
								n, err := p.node.storage.SeekRead(p.node.stateMachine.snapshotReadWriter.FileName(), p.offset, b)
								if err != nil {
									return
								}
								if uint64(n) == p.size {
									offset := p.installSnapshot(p.offset, b[:n], true)
									if offset == p.offset+uint64(n) {
										p.offset += uint64(n)
										p.chunk += 1
									}
								}
								if p.offset == p.size && p.chunk == p.chunkNum {
									p.chunk = 0
									p.offset = 0
								}
							}
						}()
					}
				}
			}

		} else {
			if p.send {
				p.send = false
				go func() {
					defer func() {
						p.send = true
					}()
					entries := p.node.log.copyAfter(p.nextIndex, DefaultMaxBatch)
					if len(entries) > 0 {
						p.appendEntries(entries)
					}
				}()
			}
		}
	}
}
