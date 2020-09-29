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
	rpc.SetLogLevel(rpc.OffLevel)
	logger.Infoln(server.Listen(network, address, codec))
}

type rpcs struct {
	*rpc.Transport
}

func newRPCs() *rpcs {
	r := &rpcs{
		&rpc.Transport{
			MaxConnsPerHost:     maxConnsPerHost,
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			Options:             &rpc.Options{Network: network, Codec: codec},
		},
	}
	return r
}

func (r *rpcs) AppendEntriesServiceName() string {
	return serviceName + "." + appendEntriesName
}

func (r *rpcs) RequestVoteServiceName() string {
	return serviceName + "." + requestVoteName
}

func (r *rpcs) InstallSnapshotServiceName() string {
	return serviceName + "." + installSnapshotName
}

func (r *rpcs) QueryLeaderServiceName() string {
	return serviceName + "." + queryLeaderName
}

func (r *rpcs) AddPeerServiceName() string {
	return serviceName + "." + addPeerName
}

func (r *rpcs) RemovePeerServiceName() string {
	return serviceName + "." + removePeerName
}

func (r *rpcs) Ping(addr string) bool {
	if err := r.Transport.Ping(addr); err != nil {
		return false
	}
	return true
}
