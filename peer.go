package raft

import(
	"sync"
)

type Peer struct {
	mu 						sync.Mutex
	node					*Node
	address					string
	alive					bool
	follower				bool
	prevLogIndex 			uint64
	prevLogTerm				uint64
	nextIndex				uint64
	matchIndex				uint64
}

func newPeer(node *Node, address string) *Peer {
	p:=&Peer{
		node :					node,
		address :				address,
	}
	return p
}

func (p *Peer) heartbeat() {
	p.alive,p.follower=p.node.raft.Heartbeat(p.address,p.prevLogIndex,p.prevLogTerm)
}

func (p *Peer) requestVote() {
	p.alive=p.node.raft.RequestVote(p.address)
}
func (p *Peer) appendEntries(entries []*Entry) {
	p.alive,_=p.node.raft.AppendEntries(p.address,p.prevLogIndex,p.prevLogTerm,entries)
}
func (p *Peer) installSnapshot() {
	p.alive=p.node.raft.InstallSnapshot(p.address)
}

func (p *Peer) ping() {
	p.alive=p.node.rpcs.Ping(p.address)
}