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
	addPeerName         = "J"
	removePeerName      = "L"
)

func listenAndServe(address string, n *node) {
	service := new(service)
	service.node = n
	server := rpc.NewServer()
	server.RegisterName(serviceName, service)
	rpc.SetLogLevel(rpc.OffLogLevel)
	logger.Infoln(server.Listen(network, address, codec))
}

type rpcs struct {
	*rpc.Transport
	appendEntriesServiceName   string
	requestVoteServiceName     string
	installSnapshotServiceName string
	queryLeaderServiceName     string
	addPeerServiceName         string
	removePeerServiceName      string
}

func newRPCs() *rpcs {
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
		addPeerServiceName:         serviceName + "." + addPeerName,
		removePeerServiceName:      serviceName + "." + removePeerName,
	}
	return r
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

func (r *rpcs) AddPeerServiceName() string {
	return r.addPeerServiceName
}

func (r *rpcs) RemovePeerServiceName() string {
	return r.removePeerServiceName
}
