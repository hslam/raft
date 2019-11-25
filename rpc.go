package raft

import (
	"hslam.com/git/x/rpc"
	"hslam.com/git/x/rpc/log"
)

const (
	MaxConnsPerHost	= 2
	MaxIdleConnsPerHost=2
	network = "tcp"
	codec = "pb"
	ServiceName = "R"
	RequestVoteName = "R"
	AppendEntriesName = "A"
	InstallSnapshotName = "I"
	QueryLeaderName = "Q"
	AddPeerName = "J"
	RemovePeerName = "L"
)

func listenAndServe(address string,node *Node){
	service:=new(RPCService)
	service.node=node
	server:= rpc.NewServer()
	server.RegisterName(ServiceName,service)
	server.EnableMultiplexing()
	server.SetLowDelay(true)
	rpc.SetLogLevel(log.NoLevel)
	Infoln(server.ListenAndServe(network,address))
}

type RPCs struct {
	conns 			*rpc.Transport
}

func newRPCs() *RPCs{
	opts:=rpc.DefaultOptions()
	opts.SetMultiplexing(true)
	opts.SetRetry(false)
	opts.SetCompressType("gzip")
	opts.SetLowDelay(true)
	r :=&RPCs{
		conns:rpc.NewTransport(MaxConnsPerHost,MaxIdleConnsPerHost,network,codec,opts),
	}
	return r
}

func (r *RPCs) AppendEntriesServiceName() string {
	return ServiceName+"."+AppendEntriesName
}

func (r *RPCs) RequestVoteServiceName() string {
	return ServiceName+"."+RequestVoteName
}

func (r *RPCs) InstallSnapshotServiceName() string {
	return ServiceName+"."+InstallSnapshotName
}

func (r *RPCs) QueryLeaderServiceName() string {
	return ServiceName+"."+QueryLeaderName
}

func (r *RPCs) AddPeerServiceName() string {
	return ServiceName+"."+AddPeerName
}

func (r *RPCs) RemovePeerServiceName() string {
	return ServiceName+"."+RemovePeerName
}

func (r *RPCs) Ping(addr string)bool {
	return r.conns.Ping(addr)
}

func (r *RPCs) Call(name string,req interface{}, res interface{},addr string)error {
	return r.conns.Call(name,req,res,addr)
}
