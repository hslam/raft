// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/rpc"
)

const (
	maxConnsPerHost     = 2
	maxIdleConnsPerHost = 2
	network             = "tcp"
	codec               = "pb"
	serviceName         = "R"
	requestVoteName     = "R"
	appendEntriesName   = "A"
	installSnapshotName = "I"
	getLeaderName       = "Q"
	addMemberName       = "J"
	removeMemberName    = "L"
	setMetaName         = "S"
	getMetaName         = "G"
)

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
	node                       *node
	server                     *rpc.Server
	addr                       string
}

func newRPCs(n *node, addr string) *rpcs {
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
		node:                       n,
		addr:                       addr,
	}
	return r
}

func (r *rpcs) ListenAndServe() {
	service := new(service)
	service.node = r.node
	r.server = rpc.NewServer()
	r.server.RegisterName(serviceName, service)
	r.server.SetLogLevel(rpc.OffLogLevel)
	r.server.SetNoCopy(true)
	r.node.logger.Errorln(r.server.Listen(network, r.addr, codec))
}

func (r *rpcs) Stop() error {
	r.Transport.Close()
	return r.server.Close()
}

func (r *rpcs) AppendEntriesServiceName() string {
	return r.appendEntriesServiceName
}

func (r *rpcs) RequestVoteServiceName() string {
	return r.requestVoteServiceName
}

func (r *rpcs) InstallSnapshotServiceName() string {
	return r.installSnapshotServiceName
}

func (r *rpcs) GetLeaderServiceName() string {
	return r.getLeaderServiceName
}

func (r *rpcs) AddMemberServiceName() string {
	return r.addMemberServiceName
}

func (r *rpcs) RemoveMemberServiceName() string {
	return r.removeMemberServiceName
}

func (r *rpcs) SetMetaServiceName() string {
	return r.setMetaServiceName
}

func (r *rpcs) GetMetaServiceName() string {
	return r.getMetaServiceName
}
