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
	service := new(RPCService)
	service.node = node
	server := rpc.NewServer()
	server.RegisterName(ServiceName, service)
	rpc.SetLogLevel(rpc.OffLevel)
	Infoln(server.Listen(network, address, codec))
}

type RPCs struct {
	conns *rpc.Transport
}

func newRPCs() *RPCs {
	r := &RPCs{
		conns: &rpc.Transport{
			MaxConnsPerHost:     MaxConnsPerHost,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			Options:             &rpc.Options{Network: network, Codec: codec},
		},
	}
	return r
}

func (r *RPCs) AppendEntriesServiceName() string {
	return ServiceName + "." + AppendEntriesName
}

func (r *RPCs) RequestVoteServiceName() string {
	return ServiceName + "." + RequestVoteName
}

func (r *RPCs) InstallSnapshotServiceName() string {
	return ServiceName + "." + InstallSnapshotName
}

func (r *RPCs) QueryLeaderServiceName() string {
	return ServiceName + "." + QueryLeaderName
}

func (r *RPCs) AddPeerServiceName() string {
	return ServiceName + "." + AddPeerName
}

func (r *RPCs) RemovePeerServiceName() string {
	return ServiceName + "." + RemovePeerName
}

func (r *RPCs) Ping(addr string) bool {
	if err := r.conns.Ping(addr); err != nil {
		return false
	}
	return true
}

func (r *RPCs) Call(addr string, name string, req interface{}, res interface{}) error {
	return r.conns.Call(addr, name, req, res)
}
