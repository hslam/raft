// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"time"
)

// Cluster represents the cluster service.
type Cluster interface {
	CallQueryLeader(addr string) (term uint64, LeaderID string, ok bool)
	CallSetPeer(addr string, info *NodeInfo) (success bool, LeaderID string, ok bool)
	CallRemovePeer(addr string, Address string) (success bool, LeaderID string, ok bool)
	CallSetMeta(addr string, meta []byte) (ok bool)
	CallGetMeta(addr string) (meta []byte, ok bool)
	QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error
	SetPeer(req *SetPeerRequest, res *SetPeerResponse) error
	RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error
	SetMeta(req *SetMetaRequest, res *SetMetaResponse) error
	GetMeta(req *GetMetaRequest, res *GetMetaResponse) error
}

type cluster struct {
	node               *node
	queryLeaderTimeout time.Duration
	addPeerTimeout     time.Duration
	removePeerTimeout  time.Duration
	setMetaTimeout     time.Duration
	getMetaTimeout     time.Duration
}

func newCluster(n *node) Cluster {
	return &cluster{
		node:               n,
		queryLeaderTimeout: defaultQueryLeaderTimeout,
		addPeerTimeout:     defaultAddPeerTimeout,
		removePeerTimeout:  defaultRemovePeerTimeout,
		setMetaTimeout:     defaultSetMetaTimeout,
		getMetaTimeout:     defaultGetMetaTimeout,
	}
}

func (c *cluster) CallQueryLeader(addr string) (term uint64, LeaderID string, ok bool) {
	var req = &QueryLeaderRequest{}
	var res = &QueryLeaderResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.QueryLeaderServiceName(), req, res, c.queryLeaderTimeout)
	if err != nil {
		c.node.logger.Tracef("raft.CallQueryLeader %s -> %s error %s", c.node.address, addr, err.Error())
		return 0, "", false
	}
	c.node.logger.Tracef("raft.CallQueryLeader %s -> %s LeaderID %s", c.node.address, addr, res.LeaderID)
	return res.Term, res.LeaderID, true
}

func (c *cluster) CallSetPeer(addr string, info *NodeInfo) (success bool, LeaderID string, ok bool) {
	var req = &SetPeerRequest{}
	req.Node = info
	var res = &SetPeerResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.SetPeerServiceName(), req, res, c.addPeerTimeout)
	if err != nil {
		c.node.logger.Tracef("raft.CallSetPeer %s -> %s error %s", c.node.address, addr, err.Error())
		return false, res.LeaderID, false
	}
	c.node.logger.Tracef("raft.CallSetPeer %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, res.LeaderID, true
}

func (c *cluster) CallRemovePeer(addr string, Address string) (success bool, LeaderID string, ok bool) {
	var req = &RemovePeerRequest{}
	req.Address = Address
	var res = &RemovePeerResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.RemovePeerServiceName(), req, res, c.removePeerTimeout)
	if err != nil {
		c.node.logger.Tracef("raft.CallRemovePeer %s -> %s error %s", c.node.address, addr, err.Error())
		return false, res.LeaderID, false
	}
	c.node.logger.Tracef("raft.CallRemovePeer %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, res.LeaderID, true
}

func (c *cluster) CallSetMeta(addr string, meta []byte) (ok bool) {
	var req = &SetMetaRequest{Meta: meta}
	var res = &SetMetaResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.SetMetaServiceName(), req, res, c.setMetaTimeout)
	if err != nil {
		c.node.logger.Tracef("raft.CallSetMeta %s -> %s error %s", c.node.address, addr, err.Error())
		return false
	}
	c.node.logger.Tracef("raft.CallSetMeta %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success
}

func (c *cluster) CallGetMeta(addr string) (meta []byte, ok bool) {
	var req = &GetMetaRequest{}
	var res = &GetMetaResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.GetMetaServiceName(), req, res, c.getMetaTimeout)
	if err != nil {
		c.node.logger.Tracef("raft.CallGetMeta %s -> %s error %s", c.node.address, addr, err.Error())
		return nil, false
	}
	c.node.logger.Tracef("raft.CallGetMeta %s -> %s Meta %v", c.node.address, addr, res.Meta)
	return res.Meta, true
}

func (c *cluster) QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	if c.node.leader.Load() != "" {
		res.LeaderID = c.node.leader.Load()
		res.Term = c.node.currentTerm.Load()
		return nil
	}
	return ErrNotLeader
}

func (c *cluster) SetPeer(req *SetPeerRequest, res *SetPeerResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newSetPeerCommand(req.Node), defaultCommandTimeout)
		if err == nil {
			_, err = c.node.do(reconfigurationCommand, defaultCommandTimeout)
			if err == nil {
				res.Success = true
			}
		}
		return err
	}
	res.LeaderID = c.node.leader.Load()
	return ErrNotLeader
}

func (c *cluster) RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newRemovePeerCommand(req.Address), defaultCommandTimeout)
		if err == nil {
			_, err = c.node.do(reconfigurationCommand, defaultCommandTimeout)
			if err == nil {
				res.Success = true
			}
		}
		return err
	}
	res.LeaderID = c.node.leader.Load()
	return ErrNotLeader
}

func (c *cluster) SetMeta(req *SetMetaRequest, res *SetMetaResponse) error {
	c.node.meta = req.Meta
	res.Success = true
	return nil
}

func (c *cluster) GetMeta(req *GetMetaRequest, res *GetMetaResponse) error {
	res.Meta = c.node.meta
	return nil
}
