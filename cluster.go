// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"context"
	"time"
)

// Cluster represents the cluster service.
type Cluster interface {
	CallGetLeader(addr string) (term uint64, LeaderID string, ok bool)
	CallAddMember(addr string, member *Member) (success bool, LeaderID string, ok bool)
	CallRemoveMember(addr string, Address string) (success bool, LeaderID string, ok bool)
	CallSetMeta(addr string, meta []byte) (ok bool)
	CallGetMeta(addr string) (meta []byte, ok bool)
	GetLeader(req *GetLeaderRequest, res *GetLeaderResponse) error
	AddMember(req *AddMemberRequest, res *AddMemberResponse) error
	RemoveMember(req *RemoveMemberRequest, res *RemoveMemberResponse) error
	SetMeta(req *SetMetaRequest, res *SetMetaResponse) error
	GetMeta(req *GetMetaRequest, res *GetMetaResponse) error
}

type cluster struct {
	node                *node
	getLeaderTimeout    time.Duration
	addMemberTimeout    time.Duration
	removeMemberTimeout time.Duration
	setMetaTimeout      time.Duration
	getMetaTimeout      time.Duration
}

func newCluster(n *node) Cluster {
	return &cluster{
		node:                n,
		getLeaderTimeout:    defaultGetLeaderTimeout,
		addMemberTimeout:    defaultAddMemberTimeout,
		removeMemberTimeout: defaultRemoveMemberTimeout,
		setMetaTimeout:      defaultSetMetaTimeout,
		getMetaTimeout:      defaultGetMetaTimeout,
	}
}

func (c *cluster) CallGetLeader(addr string) (term uint64, LeaderID string, ok bool) {
	var req = &GetLeaderRequest{}
	var res = &GetLeaderResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), c.getLeaderTimeout)
	defer cancel()
	err := c.node.rpcs.CallWithContext(ctx, addr, c.node.rpcs.GetLeaderServiceName(), req, res)
	if err != nil {
		c.node.logger.Tracef("raft.CallGetLeader %s -> %s error %s", c.node.address, addr, err.Error())
		return 0, "", false
	}
	c.node.logger.Tracef("raft.CallGetLeader %s -> %s LeaderID %s", c.node.address, addr, res.LeaderID)
	return res.Term, cloneString(res.LeaderID), true
}

func (c *cluster) CallAddMember(addr string, member *Member) (success bool, LeaderID string, ok bool) {
	var req = &AddMemberRequest{}
	req.Member = member
	var res = &AddMemberResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), c.addMemberTimeout)
	defer cancel()
	err := c.node.rpcs.CallWithContext(ctx, addr, c.node.rpcs.AddMemberServiceName(), req, res)
	if err != nil {
		c.node.logger.Tracef("raft.CallAddMember %s -> %s error %s", c.node.address, addr, err.Error())
		return false, cloneString(res.LeaderID), false
	}
	c.node.logger.Tracef("raft.CallAddMember %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, cloneString(res.LeaderID), true
}

func (c *cluster) CallRemoveMember(addr string, Address string) (success bool, LeaderID string, ok bool) {
	var req = &RemoveMemberRequest{}
	req.Address = Address
	var res = &RemoveMemberResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), c.removeMemberTimeout)
	defer cancel()
	err := c.node.rpcs.CallWithContext(ctx, addr, c.node.rpcs.RemoveMemberServiceName(), req, res)
	if err != nil {
		c.node.logger.Tracef("raft.CallRemoveMember %s -> %s error %s", c.node.address, addr, err.Error())
		return false, cloneString(res.LeaderID), false
	}
	c.node.logger.Tracef("raft.CallRemoveMember %s -> %s Success %t", c.node.address, addr, res.Success)
	return res.Success, cloneString(res.LeaderID), true
}

func (c *cluster) CallSetMeta(addr string, meta []byte) (ok bool) {
	var req = &SetMetaRequest{Meta: meta}
	var res = &SetMetaResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), c.setMetaTimeout)
	defer cancel()
	err := c.node.rpcs.CallWithContext(ctx, addr, c.node.rpcs.SetMetaServiceName(), req, res)
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
	ctx, cancel := context.WithTimeout(context.Background(), c.getMetaTimeout)
	defer cancel()
	err := c.node.rpcs.CallWithContext(ctx, addr, c.node.rpcs.GetMetaServiceName(), req, res)
	if err != nil {
		c.node.logger.Tracef("raft.CallGetMeta %s -> %s error %s", c.node.address, addr, err.Error())
		return nil, false
	}
	c.node.logger.Tracef("raft.CallGetMeta %s -> %s Meta length %d", c.node.address, addr, len(res.Meta))
	return cloneBytes(res.Meta), true
}

func (c *cluster) GetLeader(req *GetLeaderRequest, res *GetLeaderResponse) error {
	if c.node.leader.Load() != "" {
		res.LeaderID = c.node.leader.Load()
		res.Term = c.node.currentTerm.Load()
		return nil
	}
	return ErrNotLeader
}

func (c *cluster) AddMember(req *AddMemberRequest, res *AddMemberResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newAddMemberCommand(cloneString(req.Member.Address), req.Member.NonVoting), defaultCommandTimeout)
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

func (c *cluster) RemoveMember(req *RemoveMemberRequest, res *RemoveMemberResponse) error {
	if c.node.IsLeader() {
		_, err := c.node.do(newRemoveMemberCommand(cloneString(req.Address)), defaultCommandTimeout)
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
	c.node.meta = cloneBytes(req.Meta)
	res.Success = true
	return nil
}

func (c *cluster) GetMeta(req *GetMetaRequest, res *GetMetaResponse) error {
	res.Meta = c.node.meta
	return nil
}
