package raft

import(
	"sync"
	"time"
)

type Peer struct {
	mu 						sync.Mutex
	node					*Node
	address					string
	alive					bool
	appendEntriesTicker		*time.Ticker
	nextIndex				uint64
	matchIndex				uint64
	send 					bool
}

func newPeer(node *Node, address string) *Peer {
	p:=&Peer{
		node :					node,
		address :				address,
		appendEntriesTicker:	time.NewTicker(DefaultMaxDelay),
		send:					true,
	}
	go p.run()
	return p
}

func (p *Peer) heartbeat() {
	if !p.alive{
		return
	}
	var prevLogIndex,prevLogTerm uint64
	if p.nextIndex<=1{
		prevLogIndex=0
		prevLogTerm=0
	}else {
		entry:=p.node.log.lookup(p.nextIndex-1)
		if entry==nil{
			prevLogIndex=0
			prevLogTerm=0
		}else {
			prevLogIndex=p.nextIndex-1
			prevLogTerm=entry.Term
		}
	}
	//Tracef("Peer.heartbeat %s %d %d",p.address,prevLogIndex,prevLogTerm)
	nextIndex, term,success,ok:=p.node.raft.Hearbeat(p.address,prevLogIndex,prevLogTerm)
	if !ok{
		p.alive=false
	}else if nextIndex>0&&term>0&&success{
		p.nextIndex=nextIndex
	}
	//Tracef("Peer.heartbeat %s %d %d %t %t %d",p.address,nextIndex, term,success,ok,p.nextIndex)
}

func (p *Peer) requestVote() {
	if !p.alive{
		return
	}
	p.alive=p.node.raft.RequestVote(p.address)
}

func (p *Peer) appendEntries(entries []*Entry) (nextIndex uint64,term uint64,success bool,ok bool){
	var prevLogIndex,prevLogTerm uint64
	if p.nextIndex<=1{
		prevLogIndex=0
		prevLogTerm=0
	}else {
		entry:=p.node.log.lookup(p.nextIndex-1)
		if entry==nil{
			prevLogIndex=0
			prevLogTerm=0
		}else {
			prevLogIndex=p.nextIndex-1
			prevLogTerm=entry.Term
		}
	}
	//Tracef("Peer.run %s %d %d %d ",p.address,prevLogIndex,prevLogTerm,len(entries))
	nextIndex, term,success,ok=p.node.raft.AppendEntries(p.address,prevLogIndex,prevLogTerm,entries)
	return
}
func (p *Peer) installSnapshot() {
	if !p.alive{
		return
	}
	p.alive=p.node.raft.InstallSnapshot(p.address)
}

func (p *Peer) ping() {
	p.alive=p.node.rpcs.Ping(p.address)
	//Debugf("Peer.ping %s %t",p.address,p.alive)
}

func (p *Peer) run()  {
	for{
		select {
		case <-p.appendEntriesTicker.C:
			if p.node.state==nil{
				return
			}
			if p.node.state.String()==Leader&&p.send{
				p.send=false
				if p.node.lastLogIndex>p.nextIndex-1&&p.nextIndex>0{
					entries:=p.node.log.copyAfter(p.nextIndex,DefaultMaxBatch)
					if len(entries)>0{
						for i:=0;i<DefaultRetryTimes;i++{
							nextIndex,term,success,ok:=p.appendEntries(entries)
							if success&&ok{
								//p.nextIndex=entries[len(entries)-1].Index+1
								Tracef("Peer.run %s nextIndex %d==>%d",p.address,p.nextIndex,nextIndex)
								p.nextIndex=nextIndex
								break
							}else if ok&&term==p.node.currentTerm.Id(){
								p.nextIndex=nextIndex
								break
							}else if !ok{
								p.alive=false
							}
						}
					}

				}
				p.send=true
			}
		}
	}
}