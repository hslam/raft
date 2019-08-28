package raft

import (
	"sync"
	"time"
)

type Node struct {
	sync.RWMutex
	server				*Server
	address				string
	hearbeatTicker		time.Duration
	heartbeated			bool
	stopHearbeat		chan bool
}

func newNode(server *Server, address string) *Node {
	return &Node{
		server :				server,
		address :				address,
		hearbeatTicker:			DefaultHearbeatTicker,
		heartbeated:			false,
		stopHearbeat:			make(chan bool),
	}
}

func (n *Node) setHearbeatTicker(ticker time.Duration ) {
	n.hearbeatTicker=ticker
}
func (n *Node) startHeartbeat() {
	if n.heartbeated {
		return
	}
	n.heartbeated = true
	ticker:=time.NewTicker(n.hearbeatTicker)
	go func() {
		for{
			select {
			case <-n.stopHearbeat:
				goto endfor
			case <-ticker.C:
				go n.heartbeat()
			}
		}
		endfor:
	}()
}
func (n *Node) stopHeartbeat() {
	<-n.stopHearbeat
	n.heartbeated=false
}

func (n *Node) heartbeat() {
	var req =&AppendEntriesRequest{}
	req.Term=1
	var ch =make(chan *AppendEntriesResponse)
	go n.AppendEntries(req,ch)
	select {
	case <-ch:
		Tracef("Node.heartbeat %s -> %s",n.server.address,n.address)
	case <-time.After(n.hearbeatTicker):
		Tracef("Node.heartbeat %s -> %s time out",n.server.address,n.address)
	}
}
func (n *Node) RequestVote(req *RequestVoteRequest, ch chan *RequestVoteResponse) {
	req.CandidateId = n.address
	var res =&RequestVoteResponse{}
	if err := n.server.rpcs.CallRequestVote(n.address,req, res); err != nil {
	}else {
		ch <- res
	}
}

func (n *Node) AppendEntries(req *AppendEntriesRequest, ch chan *AppendEntriesResponse) {
	var res =&AppendEntriesResponse{}
	if err := n.server.rpcs.CallAppendEntries(n.address,req, res); err != nil {
	} else {
		ch <- res
	}
}
func (n *Node) InstallSnapshot(req *InstallSnapshotRequest, ch chan *InstallSnapshotResponse) {
	var res =&InstallSnapshotResponse{}
	if err := n.server.rpcs.CallInstallSnapshot(n.address,req, res); err != nil {
	}else {
		ch <- res
	}
}

