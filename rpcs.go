// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"context"
	"github.com/hslam/rpc"
)

const (
	maxConnsPerHost     = 2
	maxIdleConnsPerHost = 2
	network             = "tcp"
	codec               = "pb"
	serviceName         = "R"
	requestVoteName     = "RequestVote"
	appendEntriesName   = "AppendEntries"
	installSnapshotName = "InstallSnapshot"
	getLeaderName       = "GetLeader"
	addMemberName       = "AddMember"
	removeMemberName    = "RemoveMember"
	setMetaName         = "SetMeta"
	getMetaName         = "GetMeta"
)

// RPCs represents the RPCs.
type RPCs interface {
	Register(s Service) error
	ListenAndServe() error
	Close() error
	Ping(addr string) error
	RequestVote(ctx context.Context, addr string, req *RequestVoteRequest, res *RequestVoteResponse) error
	AppendEntries(ctx context.Context, addr string, req *AppendEntriesRequest, res *AppendEntriesResponse) error
	InstallSnapshot(ctx context.Context, addr string, req *InstallSnapshotRequest, res *InstallSnapshotResponse) error
	GetLeader(ctx context.Context, addr string, req *GetLeaderRequest, res *GetLeaderResponse) error
	AddMember(ctx context.Context, addr string, req *AddMemberRequest, res *AddMemberResponse) error
	RemoveMember(ctx context.Context, addr string, req *RemoveMemberRequest, res *RemoveMemberResponse) error
	SetMeta(ctx context.Context, addr string, req *SetMetaRequest, res *SetMetaResponse) error
	GetMeta(ctx context.Context, addr string, req *GetMetaRequest, res *GetMetaResponse) error
}

type rpcs struct {
	*rpc.Transport
	appendEntriesServiceName   string
	requestVoteServiceName     string
	installSnapshotServiceName string
	getLeaderServiceName       string
	addMemberServiceName       string
	removeMemberServiceName    string
	setMetaServiceName         string
	getMetaServiceName         string
	server                     *rpc.Server
	addr                       string
}

func newRPCs(addr string) *rpcs {
	r := &rpcs{
		Transport: &rpc.Transport{
			MaxConnsPerHost:     maxConnsPerHost,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			Options:             &rpc.Options{Network: network, Codec: codec},
		},
		appendEntriesServiceName:   serviceName + "." + appendEntriesName,
		requestVoteServiceName:     serviceName + "." + requestVoteName,
		installSnapshotServiceName: serviceName + "." + installSnapshotName,
		getLeaderServiceName:       serviceName + "." + getLeaderName,
		addMemberServiceName:       serviceName + "." + addMemberName,
		removeMemberServiceName:    serviceName + "." + removeMemberName,
		setMetaServiceName:         serviceName + "." + setMetaName,
		getMetaServiceName:         serviceName + "." + getMetaName,
		server:                     rpc.NewServer(),
		addr:                       addr,
	}
	return r
}

func (r *rpcs) Register(service Service) error {
	return r.server.RegisterName(serviceName, service)
}

func (r *rpcs) ListenAndServe() error {
	r.server.SetLogLevel(rpc.OffLogLevel)
	r.server.SetNoCopy(true)
	return r.server.Listen(network, r.addr, codec)
}

func (r *rpcs) Close() error {
	r.Transport.Close()
	return r.server.Close()
}

func (r *rpcs) RequestVote(ctx context.Context, addr string, req *RequestVoteRequest, res *RequestVoteResponse) error {
	return r.CallWithContext(ctx, addr, r.requestVoteServiceName, req, res)
}

func (r *rpcs) AppendEntries(ctx context.Context, addr string, req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	return r.CallWithContext(ctx, addr, r.appendEntriesServiceName, req, res)
}

func (r *rpcs) InstallSnapshot(ctx context.Context, addr string, req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	return r.CallWithContext(ctx, addr, r.installSnapshotServiceName, req, res)
}

func (r *rpcs) GetLeader(ctx context.Context, addr string, req *GetLeaderRequest, res *GetLeaderResponse) error {
	return r.CallWithContext(ctx, addr, r.getLeaderServiceName, req, res)
}

func (r *rpcs) AddMember(ctx context.Context, addr string, req *AddMemberRequest, res *AddMemberResponse) error {
	return r.CallWithContext(ctx, addr, r.addMemberServiceName, req, res)
}

func (r *rpcs) RemoveMember(ctx context.Context, addr string, req *RemoveMemberRequest, res *RemoveMemberResponse) error {
	return r.CallWithContext(ctx, addr, r.removeMemberServiceName, req, res)
}

func (r *rpcs) SetMeta(ctx context.Context, addr string, req *SetMetaRequest, res *SetMetaResponse) error {
	return r.CallWithContext(ctx, addr, r.setMetaServiceName, req, res)
}

func (r *rpcs) GetMeta(ctx context.Context, addr string, req *GetMetaRequest, res *GetMetaResponse) error {
	return r.CallWithContext(ctx, addr, r.getMetaServiceName, req, res)
}
