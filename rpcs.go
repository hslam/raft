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
	queryLeaderName     = "Q"
	setPeerName         = "J"
	removePeerName      = "L"
)

type rpcs struct {
	*rpc.Transport
	appendEntriesServiceName   string
	requestVoteServiceName     string
	installSnapshotServiceName string
	queryLeaderServiceName     string
	setPeerServiceName         string
	removePeerServiceName      string
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
		queryLeaderServiceName:     serviceName + "." + queryLeaderName,
		setPeerServiceName:         serviceName + "." + setPeerName,
		removePeerServiceName:      serviceName + "." + removePeerName,
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
	rpc.SetLogLevel(rpc.OffLogLevel)
	logger.Errorln(r.server.Listen(network, r.addr, codec))
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

func (r *rpcs) QueryLeaderServiceName() string {
	return r.queryLeaderServiceName
}

func (r *rpcs) SetPeerServiceName() string {
	return r.setPeerServiceName
}

func (r *rpcs) RemovePeerServiceName() string {
	return r.removePeerServiceName
}
