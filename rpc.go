package raft

import (
	"hslam.com/git/x/rpc"
	"hslam.com/git/x/rpc/log"
	"time"
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
type Client struct {
	rpc.Client
	keepAlive				time.Duration
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

func (r *RPCs) ServiceAppendEntriesName() string {
	return ServiceName+"."+AppendEntriesName
}

func (r *RPCs) ServiceRequestVoteName() string {
	return ServiceName+"."+RequestVoteName
}

func (r *RPCs) ServiceInstallSnapshotName() string {
	return ServiceName+"."+InstallSnapshotName
}

func (r *RPCs) Ping(addr string)bool {
	return r.conns.Ping(addr)
}

func (r *RPCs) CallRequestVote(addr string,req *RequestVoteRequest, res *RequestVoteResponse)error {
	return r.conns.Call(r.ServiceRequestVoteName(),req,res,addr)
}

func (r *RPCs) CallAppendEntries(addr string,req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	return r.conns.Call(r.ServiceAppendEntriesName(),req,res,addr)
}

func (r *RPCs) CallInstallSnapshot(addr string,req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return r.conns.Call(r.ServiceInstallSnapshotName(),req,res,addr)
}