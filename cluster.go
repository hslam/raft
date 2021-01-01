// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/rpc"
	"runtime"
	"sync"
	"time"
)

// Cluster represents the cluster service.
type Cluster interface {
	CallQueryLeader(addr string) (term uint64, leaderID string, ok bool)
	CallAddPeer(addr string, info *NodeInfo) (success bool, ok bool)
	CallRemovePeer(addr string, Address string) (success bool, ok bool)
	QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error
	AddPeer(req *AddPeerRequest, res *AddPeerResponse) error
	RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error
}

type cluster struct {
	node               *node
	callPool           *sync.Pool
	donePool           *sync.Pool
	queryLeaderTimeout time.Duration
	addPeerTimeout     time.Duration
	removePeerTimeout  time.Duration
}

func newCluster(n *node) Cluster {
	return &cluster{
		node:               n,
		callPool:           &sync.Pool{New: func() interface{} { return &rpc.Call{} }},
		donePool:           &sync.Pool{New: func() interface{} { return make(chan *rpc.Call, 10) }},
		queryLeaderTimeout: defaultQueryLeaderTimeout,
		addPeerTimeout:     defaultAddPeerTimeout,
		removePeerTimeout:  defaultRemovePeerTimeout,
	}
}

func (c *cluster) CallQueryLeader(addr string) (term uint64, leaderID string, ok bool) {
	var req = &QueryLeaderRequest{}
	done := c.donePool.Get().(chan *rpc.Call)
	call := c.callPool.Get().(*rpc.Call)
	call.ServiceMethod = c.node.rpcs.QueryLeaderServiceName()
	call.Args = req
	call.Reply = &QueryLeaderResponse{}
	call.Done = done
	c.node.rpcs.RoundTrip(addr, call)
	timer := time.NewTimer(c.queryLeaderTimeout)
	runtime.Gosched()
	select {
	case call := <-done:
		timer.Stop()
		for len(done) > 0 {
			<-done
		}
		c.donePool.Put(done)
		if call.Error != nil {
			logger.Tracef("raft.CallQueryLeader %s -> %s error %s", c.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			c.callPool.Put(call)
			return 0, "", false
		}
		res := call.Reply.(*QueryLeaderResponse)
		*call = rpc.Call{}
		c.callPool.Put(call)
		logger.Tracef("raft.CallQueryLeader %s -> %s LeaderId %s", c.node.address, addr, res.LeaderId)
		return res.Term, res.LeaderId, true
	case <-timer.C:
		logger.Tracef("raft.CallQueryLeader %s -> %s time out", c.node.address, addr)
	}
	return 0, "", false
}

func (c *cluster) CallAddPeer(addr string, info *NodeInfo) (success bool, ok bool) {
	var req = &AddPeerRequest{}
	req.Node = info
	done := c.donePool.Get().(chan *rpc.Call)
	call := c.callPool.Get().(*rpc.Call)
	call.ServiceMethod = c.node.rpcs.AddPeerServiceName()
	call.Args = req
	call.Reply = &AddPeerResponse{}
	call.Done = done
	c.node.rpcs.RoundTrip(addr, call)
	timer := time.NewTimer(c.addPeerTimeout)
	runtime.Gosched()
	select {
	case call := <-done:
		timer.Stop()
		for len(done) > 0 {
			<-done
		}
		c.donePool.Put(done)
		if call.Error != nil {
			logger.Tracef("raft.CallAddPeer %s -> %s error %s", c.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			c.callPool.Put(call)
			return false, false
		}
		res := call.Reply.(*AddPeerResponse)
		*call = rpc.Call{}
		c.callPool.Put(call)
		logger.Tracef("raft.CallAddPeer %s -> %s Success %t", c.node.address, addr, res.Success)
		return res.Success, true
	case <-timer.C:
		logger.Tracef("raft.CallAddPeer %s -> %s time out", c.node.address, addr)
	}
	return false, false
}

func (c *cluster) CallRemovePeer(addr string, Address string) (success bool, ok bool) {
	var req = &RemovePeerRequest{}
	req.Address = Address
	done := c.donePool.Get().(chan *rpc.Call)
	call := c.callPool.Get().(*rpc.Call)
	call.ServiceMethod = c.node.rpcs.RemovePeerServiceName()
	call.Args = req
	call.Reply = &RemovePeerResponse{}
	call.Done = done
	c.node.rpcs.RoundTrip(addr, call)
	timer := time.NewTimer(c.removePeerTimeout)
	runtime.Gosched()
	select {
	case call := <-done:
		timer.Stop()
		for len(done) > 0 {
			<-done
		}
		c.donePool.Put(done)
		if call.Error != nil {
			logger.Tracef("raft.CallRemovePeer %s -> %s error %s", c.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			c.callPool.Put(call)
			return false, false
		}
		res := call.Reply.(*RemovePeerResponse)
		*call = rpc.Call{}
		c.callPool.Put(call)
		logger.Tracef("raft.CallRemovePeer %s -> %s Success %t", c.node.address, addr, res.Success)
		return res.Success, true
	case <-timer.C:
		logger.Tracef("raft.CallRemovePeer %s -> %s time out", c.node.address, addr)
	}
	return false, false
}

func (c *cluster) QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	if c.node.leader != "" {
		res.LeaderId = c.node.leader
		res.Term = c.node.term()
		return nil
	}
	return ErrNotLeader
}

func (c *cluster) AddPeer(req *AddPeerRequest, res *AddPeerResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newAddPeerCommand(req.Node), defaultCommandTimeout)
		if err == nil {
			_, err = c.node.do(newReconfigurationCommand(), defaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	}
	return ErrNotLeader
}

func (c *cluster) RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newRemovePeerCommand(req.Address), defaultCommandTimeout)
		if err == nil {
			_, err = c.node.do(newReconfigurationCommand(), defaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	}
	return ErrNotLeader
}
