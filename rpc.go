// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"github.com/hslam/rpc"
)

const (
	MaxConnsPerHost     = 2
	MaxIdleConnsPerHost = 2
	network             = "tcp"
	codec               = "pb"
	ServiceName         = "R"
	RequestVoteName     = "R"
	AppendEntriesName   = "A"
	InstallSnapshotName = "I"
	QueryLeaderName     = "Q"
	AddPeerName         = "J"
	RemovePeerName      = "L"
)

func listenAndServe(address string, node *Node) {
	service := new(service)
	service.node = node
	server := rpc.NewServer()
	server.RegisterName(ServiceName, service)
	rpc.SetLogLevel(rpc.OffLevel)
	Infoln(server.Listen(network, address, codec))
}

type rpcs struct {
	*rpc.Transport
}

func newRPCs() *rpcs {
	r := &rpcs{
		&rpc.Transport{
			MaxConnsPerHost:     MaxConnsPerHost,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			Options:             &rpc.Options{Network: network, Codec: codec},
		},
	}
	return r
}

func (r *rpcs) AppendEntriesServiceName() string {
	return ServiceName + "." + AppendEntriesName
}

func (r *rpcs) RequestVoteServiceName() string {
	return ServiceName + "." + RequestVoteName
}

func (r *rpcs) InstallSnapshotServiceName() string {
	return ServiceName + "." + InstallSnapshotName
}

func (r *rpcs) QueryLeaderServiceName() string {
	return ServiceName + "." + QueryLeaderName
}

func (r *rpcs) AddPeerServiceName() string {
	return ServiceName + "." + AddPeerName
}

func (r *rpcs) RemovePeerServiceName() string {
	return ServiceName + "." + RemovePeerName
}

func (r *rpcs) Ping(addr string) bool {
	if err := r.Transport.Ping(addr); err != nil {
		return false
	}
	return true
}
