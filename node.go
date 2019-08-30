package raft

import (
	"sync"
	"time"
)

type Node struct {
	mu sync.RWMutex
	server					*Server
	address					string


	hearbeatTimeout			time.Duration
	requestVoteTimeout		time.Duration
	appendEntriesTimeout	time.Duration
	installSnapshotTimeout	time.Duration

	alive					bool

}

func newNode(server *Server, address string) *Node {
	n:=&Node{
		server :				server,
		address :				address,
		hearbeatTimeout:		DefaultHearbeatTimeout,
		requestVoteTimeout:		DefaultRequestVoteTimeout,
		appendEntriesTimeout:	DefaultAppendEntriesTimeout,
		installSnapshotTimeout:	DefaultInstallSnapshotTimeout,
	}
	return n
}

func (n *Node) heartbeat() {
	var req =&AppendEntriesRequest{}
	req.Term=n.server.currentTerm
	req.LeaderId=n.server.leader
	req.LeaderCommit=n.server.commitIndex
	req.PrevLogIndex=n.server.commitIndex
	req.PrevLogTerm=n.server.currentTerm
	req.Entries=[]*Entry{}
	var ch =make(chan *AppendEntriesResponse)
	go n.callAppendEntries(req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("Node.heartbeat %s -> %s error",n.server.address,n.address)
			n.alive=false
			return
		}
		n.alive=true
		//Tracef("Node.heartbeat %s -> %s",n.server.address,n.address)
	case <-time.After(n.hearbeatTimeout):
		n.alive=false
		Tracef("Node.heartbeat %s -> %s time out",n.server.address,n.address)
	}
}

func (n *Node) requestVote() {
	var req =&RequestVoteRequest{}
	req.Term=n.server.currentTerm+1
	req.CandidateId=n.server.address
	req.LastLogIndex=n.server.commitIndex
	req.LastLogTerm=n.server.currentTerm
	var ch =make(chan *RequestVoteResponse)
	go n.callRequestVote(req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("Node.requestVote %s recv %s vote error",n.server.address,n.address)
			n.alive=false
			return
		}
		n.alive=true
		Tracef("Node.requestVote %s recv %s vote %t",n.server.address,n.address,res.VoteGranted)
		if res.VoteGranted{
			n.server.votes<-&Vote{candidateId:n.address,vote:1,term:req.Term}
		}else {
			n.server.votes<-&Vote{candidateId:n.address,vote:0,term:req.Term}
			if res.Term>n.server.currentTerm{
				n.server.currentTerm=res.Term
			}
		}

	case <-time.After(n.requestVoteTimeout):
		n.alive=false
		Tracef("Node.requestVote %s recv %s vote time out",n.server.address,n.address)
		n.server.votes<-&Vote{candidateId:n.address,vote:0,term:req.Term}
	}
}
func (n *Node) appendEntries() {
	var req =&AppendEntriesRequest{}
	req.Term=1
	var ch =make(chan *AppendEntriesResponse)
	go n.callAppendEntries(req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("Node.appendEntries %s -> %s error",n.server.address,n.address)
			n.alive=false
			return
		}
		n.alive=true
		Tracef("Node.appendEntries %s -> %s",n.server.address,n.address)
	case <-time.After(n.appendEntriesTimeout):
		n.alive=false
		Tracef("Node.appendEntries %s -> %s time out",n.server.address,n.address)
	}
}
func (n *Node) installSnapshot() {
	var req =&InstallSnapshotRequest{}
	req.Term=1
	var ch =make(chan *InstallSnapshotResponse)
	go n.callInstallSnapshot(req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("Node.installSnapshot %s -> %s error",n.server.address,n.address)
			n.alive=false
			return
		}
		n.alive=true
		Tracef("Node.installSnapshot %s -> %s",n.server.address,n.address)
	case <-time.After(n.installSnapshotTimeout):
		n.alive=false
		Tracef("Node.installSnapshot %s -> %s time out",n.server.address,n.address)
	}
}

func (n *Node) ping() {
	if n.server.rpcs.Ping(n.address){
		//Tracef("Node.ping %s -> %s true",n.server.address,n.address)
		n.alive=true
	}else {
		//Tracef("Node.ping %s -> %s false",n.server.address,n.address)
		n.alive=false
	}
}
func (n *Node) callRequestVote(req *RequestVoteRequest, ch chan *RequestVoteResponse) {
	var res =&RequestVoteResponse{}
	if err := n.server.rpcs.CallRequestVote(n.address,req, res); err != nil {
		ch<-nil
	}else {
		ch <- res
	}
}

func (n *Node) callAppendEntries(req *AppendEntriesRequest, ch chan *AppendEntriesResponse) {
	var res =&AppendEntriesResponse{}
	if err := n.server.rpcs.CallAppendEntries(n.address,req, res); err != nil {
		ch<-nil
	} else {
		ch <- res
	}
}
func (n *Node) callInstallSnapshot(req *InstallSnapshotRequest, ch chan *InstallSnapshotResponse) {
	var res =&InstallSnapshotResponse{}
	if err := n.server.rpcs.CallInstallSnapshot(n.address,req, res); err != nil {
		ch<-nil
	}else {
		ch <- res
	}
}

