// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
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
	queryLeaderTimeout time.Duration
	addPeerTimeout     time.Duration
	removePeerTimeout  time.Duration
}

func newCluster(n *node) Cluster {
	return &cluster{
		node:               n,
		queryLeaderTimeout: defaultQueryLeaderTimeout,
		addPeerTimeout:     defaultAddPeerTimeout,
		removePeerTimeout:  defaultRemovePeerTimeout,
	}
}

func (c *cluster) CallQueryLeader(addr string) (term uint64, leaderID string, ok bool) {
	var req = &QueryLeaderRequest{}
	var res = &QueryLeaderResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.QueryLeaderServiceName(), req, res, c.queryLeaderTimeout)
	if err != nil {
		logger.Tracef("raft.CallQueryLeader %s -> %s error %s", c.node.address, addr, err.Error())
		return 0, "", false
	}
	logger.Tracef("raft.CallQueryLeader %s -> %s LeaderId %s", c.node.address, addr, res.LeaderId)
	return res.Term, res.LeaderId, true
}

func (c *cluster) CallAddPeer(addr string, info *NodeInfo) (success bool, ok bool) {
	var req = &AddPeerRequest{}
	req.Node = info
	var res = &AddPeerResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.AddPeerServiceName(), req, res, c.addPeerTimeout)
	if err != nil {
		logger.Tracef("raft.CallAddPeer %s -> %s error %s", c.node.address, addr, err.Error())
		return false, false
	}
	logger.Tracef("raft.CallAddPeer %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, true
}

func (c *cluster) CallRemovePeer(addr string, Address string) (success bool, ok bool) {
	var req = &RemovePeerRequest{}
	req.Address = Address
	var res = &RemovePeerResponse{}
	err := c.node.rpcs.CallTimeout(addr, c.node.rpcs.RemovePeerServiceName(), req, res, c.removePeerTimeout)
	if err != nil {
		logger.Tracef("raft.CallRemovePeer %s -> %s error %s", c.node.address, addr, err.Error())
		return false, false
	}
	logger.Tracef("raft.CallRemovePeer %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, true
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
