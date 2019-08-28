package raft

import (
	"hslam.com/mgit/Mort/rpc"
	"hslam.com/mgit/Mort/rpc/log"
	"errors"
)

const (
	network = "tcp"
	codec = "pb"
	ServiceName = "R"
	RequestVoteName = "R"
	AppendEntriesName = "A"
	InstallSnapshotName = "I"

	RetryTimes = 5

)

func listenAndServe(address string){
	rpc.RegisterName(ServiceName,new(Service))
	rpc.SetLogLevel(log.ErrorLevel)
	go func() {
		Infoln(rpc.ListenAndServe(network,address))
	}()
}

type RPCs struct {
	conns map[string]*rpc.Client
	address string
}
func newRPCs(address string,addrs []string) *RPCs{
	c :=&RPCs{
		conns:make(map[string]*rpc.Client),
		address:address,
	}
	for _, addr:= range addrs {
		conn, err := c.NewConn(addr)
		if err==nil{
			c.conns[addr] = conn
		}
	}
	return c
}

func (r *RPCs) GetConn(addr string) *rpc.Client {
	if _,ok:=r.conns[addr];ok{
		return r.conns[addr]
	}
	for i:=0;i<RetryTimes;i++{
		conn, err := r.NewConn(addr)
		if err==nil{
			r.conns[addr] = conn
			return r.conns[addr]
		}
	}

	return nil
}
func (r *RPCs) NewConn(addr string) (*rpc.Client, error){
	return rpc.Dial(network,addr,codec)
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

func (r *RPCs) CallRequestVote(addr string,req *RequestVoteRequest, res *RequestVoteResponse)error {
	Tracef("RPCs.CallRequestVote %s -> %s",r.address,addr)
	conn:= r.GetConn(addr)
	if conn!=nil{
		return conn.Call(r.ServiceRequestVoteName(),req,res)
	}
	return errors.New("RPCs.CallRequestVote can not connect to "+addr)
}
func (r *RPCs) CallAppendEntries(addr string,req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	Tracef("RPCs.CallAppendEntries %s -> %s ",r.address,addr)
	return r.GetConn(addr).Call(r.ServiceAppendEntriesName(),req,res)
}
func (r *RPCs) CallInstallSnapshot(addr string,req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	Tracef("RPCs.CallAppendEntries %s -> %s",r.address,addr)
	return r.GetConn(addr).Call(r.ServiceInstallSnapshotName(),req,res)
}