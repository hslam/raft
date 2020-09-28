// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/rpc"
	"sync"
	"time"
)

// Proxy represents the proxy service.
type Proxy interface {
	QueryLeader(addr string) (term uint64, leaderID string, ok bool)
	AddPeer(addr string, info *NodeInfo) (success bool, ok bool)
	RemovePeer(addr string, Address string) (success bool, ok bool)
	HandleQueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error
	HandleAddPeer(req *AddPeerRequest, res *AddPeerResponse) error
	HandleRemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error
}

type proxy struct {
	node               *node
	callPool           *sync.Pool
	donePool           *sync.Pool
	queryLeaderTimeout time.Duration
	addPeerTimeout     time.Duration
	removePeerTimeout  time.Duration
}

func newProxy(n *node) Proxy {
	return &proxy{
		node:               n,
		callPool:           &sync.Pool{New: func() interface{} { return &rpc.Call{} }},
		donePool:           &sync.Pool{New: func() interface{} { return make(chan *rpc.Call, 10) }},
		queryLeaderTimeout: defaultQueryLeaderTimeout,
		addPeerTimeout:     defaultAddPeerTimeout,
		removePeerTimeout:  defaultRemovePeerTimeout,
	}
}

func (p *proxy) QueryLeader(addr string) (term uint64, leaderID string, ok bool) {
	var req = &QueryLeaderRequest{}
	done := p.donePool.Get().(chan *rpc.Call)
	call := p.callPool.Get().(*rpc.Call)
	call.ServiceMethod = p.node.rpcs.QueryLeaderServiceName()
	call.Args = req
	call.Reply = &QueryLeaderResponse{}
	call.Done = done
	p.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		p.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.QueryLeader %s -> %s error %s", p.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			p.callPool.Put(call)
			return 0, "", false
		}
		res := call.Reply.(*QueryLeaderResponse)
		*call = rpc.Call{}
		p.callPool.Put(call)
		Tracef("raft.QueryLeader %s -> %s LeaderId %s", p.node.address, addr, res.LeaderId)
		return res.Term, res.LeaderId, true
	case <-time.After(p.addPeerTimeout):
		Tracef("raft.QueryLeader %s -> %s time out", p.node.address, addr)
	}
	return 0, "", false
}
func (p *proxy) AddPeer(addr string, info *NodeInfo) (success bool, ok bool) {
	var req = &AddPeerRequest{}
	req.Node = info
	done := p.donePool.Get().(chan *rpc.Call)
	call := p.callPool.Get().(*rpc.Call)
	call.ServiceMethod = p.node.rpcs.AddPeerServiceName()
	call.Args = req
	call.Reply = &AddPeerResponse{}
	call.Done = done
	p.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		p.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.AddPeer %s -> %s error %s", p.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			p.callPool.Put(call)
			return false, false
		}
		res := call.Reply.(*AddPeerResponse)
		*call = rpc.Call{}
		p.callPool.Put(call)
		Tracef("raft.AddPeer %s -> %s Success %t", p.node.address, addr, res.Success)
		return res.Success, true
	case <-time.After(p.addPeerTimeout):
		Tracef("raft.AddPeer %s -> %s time out", p.node.address, addr)
	}
	return false, false
}
func (p *proxy) RemovePeer(addr string, Address string) (success bool, ok bool) {
	var req = &RemovePeerRequest{}
	req.Address = Address
	done := p.donePool.Get().(chan *rpc.Call)
	call := p.callPool.Get().(*rpc.Call)
	call.ServiceMethod = p.node.rpcs.RemovePeerServiceName()
	call.Args = req
	call.Reply = &RemovePeerResponse{}
	call.Done = done
	p.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		p.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.RemovePeer %s -> %s error %s", p.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			p.callPool.Put(call)
			return false, false
		}
		res := call.Reply.(*RemovePeerResponse)
		*call = rpc.Call{}
		p.callPool.Put(call)
		Tracef("raft.RemovePeer %s -> %s Success %t", p.node.address, addr, res.Success)
		return res.Success, true
	case <-time.After(p.removePeerTimeout):
		Tracef("raft.RemovePeer %s -> %s time out", p.node.address, addr)
	}
	return false, false
}
func (p *proxy) HandleQueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	if p.node.leader != "" {
		res.LeaderId = p.node.leader
		res.Term = p.node.term()
		return nil
	}
	peers := p.node.Peers()
	for i := 0; i < len(peers); i++ {
		term, leaderID, ok := p.QueryLeader(peers[i])
		if ok {
			res.LeaderId = leaderID
			res.Term = term
			return nil
		}
	}
	return ErrNotLeader
}
func (p *proxy) HandleAddPeer(req *AddPeerRequest, res *AddPeerResponse) error {
	if p.node.IsLeader() {
		_, err := p.node.do(NewAddPeerCommand(req.Node), defaultCommandTimeout)
		if err == nil {
			_, err = p.node.do(NewReconfigurationCommand(), defaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	}
	leader := p.node.Leader()
	if leader != "" {
		return p.node.rpcs.Call(leader, p.node.rpcs.AddPeerServiceName(), req, res)
	}
	peers := p.node.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := p.QueryLeader(peers[i])
		if leaderID != "" && ok {
			return p.node.rpcs.Call(leaderID, p.node.rpcs.AddPeerServiceName(), req, res)
		}
	}
	return ErrNotLeader
}

func (p *proxy) HandleRemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	if p.node.IsLeader() {
		_, err := p.node.do(NewRemovePeerCommand(req.Address), defaultCommandTimeout)
		if err == nil {
			_, err = p.node.do(NewReconfigurationCommand(), defaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	}
	leader := p.node.Leader()
	if leader != "" {
		return p.node.rpcs.Call(leader, p.node.rpcs.RemovePeerServiceName(), req, res)
	}
	peers := p.node.Peers()
	for i := 0; i < len(peers); i++ {
		_, leaderID, ok := p.QueryLeader(peers[i])
		if leaderID != "" && ok {
			return p.node.rpcs.Call(leaderID, p.node.rpcs.RemovePeerServiceName(), req, res)
		}
	}
	return ErrNotLeader
}
