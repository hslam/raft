package raft

import (
	"hslam.com/mgit/Mort/rpc"
	"hslam.com/mgit/Mort/rpc/log"
	"errors"
	"sync"
)

const (
	network = "tcp"
	codec = "pb"
	ServiceName = "R"
	RequestVoteName = "R"
	AppendEntriesName = "A"
	InstallSnapshotName = "I"

	RetryTimes = 0

)

func listenAndServe(address string,node *Node){
	service:=new(RPCService)
	service.node=node
	server:= rpc.NewServer()
	server.RegisterName(ServiceName,service)
	server.EnableMultiplexing()
	rpc.SetLogLevel(log.NoLevel)
	Infoln(server.ListenAndServe(network,address))
}

type RPCs struct {
	mu				sync.RWMutex
	conns			map[string]rpc.Client
}
func newRPCs(addrs []string) *RPCs{
	r :=&RPCs{
		conns:make(map[string]rpc.Client),
	}
	for _, addr:= range addrs {
		conn, err := r.NewConn(addr)
		if err==nil{
			conn.DisableRetry()
			conn.SetCompressType("gzip")
			conn.EnableMultiplexing()
			r.conns[addr] = conn
		}
	}
	return r
}

func (r *RPCs) GetConn(addr string) rpc.Client {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _,ok:=r.conns[addr];ok{
		return r.conns[addr]
	}
	conn, err := r.NewConn(addr)
	if err==nil{
		conn.DisableRetry()
		conn.SetCompressType("gzip")
		conn.EnableMultiplexing()
		r.conns[addr] = conn
		return r.conns[addr]
	}
	return nil
}

func (r *RPCs) RemoveConn(addr string){
	r.mu.Lock()
	defer r.mu.Unlock()
	//Debugf("RPCs.RemoveConn %s",addr)
	if conn,ok:=r.conns[addr];ok{
		delete(r.conns,addr)
		defer conn.Close()
	}
}
func (r *RPCs) NewConn(addr string) (rpc.Client, error){
	//Debugf("RPCs.NewConn %s",addr)
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
func (r *RPCs) Ping(addr string)bool {
	conn:= r.GetConn(addr)
	if conn!=nil{
		if conn.Ping(){
			return true
		}
	}
	r.RemoveConn(addr)
	return false
}
//func (r *RPCs) CallRequestVote(addr string,req *RequestVoteRequest, res *RequestVoteResponse)error {
//	conn:= r.GetConn(addr)
//	if conn!=nil{
//		return conn.Call(r.ServiceRequestVoteName(),req,res)
//	}
//	return errors.New("RPCs.CallRequestVote can not connect to "+addr)
//}
//func (r *RPCs) CallAppendEntries(addr string,req *AppendEntriesRequest, res *AppendEntriesResponse)error {
//	conn:= r.GetConn(addr)
//	if conn!=nil{
//		return conn.Call(r.ServiceAppendEntriesName(),req,res)
//	}
//	return errors.New("RPCs.CallAppendEntries can not connect to "+addr)
//}
//func (r *RPCs) CallInstallSnapshot(addr string,req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
//	conn:= r.GetConn(addr)
//	if conn!=nil{
//		return conn.Call(r.ServiceInstallSnapshotName(),req,res)
//	}
//	return errors.New("RPCs.CallInstallSnapshot can not connect to "+addr)
//}

func (r *RPCs) CallRequestVote(addr string,req *RequestVoteRequest, res *RequestVoteResponse)error {
	conn:= r.GetConn(addr)
	if conn!=nil{
		err:=conn.Call(r.ServiceRequestVoteName(),req,res)
		if err!=nil{
			r.RemoveConn(addr)
			return err
		}
		return nil
	}
	return errors.New("RPCs.CallRequestVote can not connect to "+addr)
}
func (r *RPCs) CallAppendEntries(addr string,req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	conn:= r.GetConn(addr)
	if conn!=nil{
		err:= conn.Call(r.ServiceAppendEntriesName(),req,res)
		if err!=nil{
			r.RemoveConn(addr)
			return err
		}
		return nil
	}
	return errors.New("RPCs.CallAppendEntries can not connect to "+addr)
}
func (r *RPCs) CallInstallSnapshot(addr string,req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	conn:= r.GetConn(addr)
	if conn!=nil{
		err:= conn.Call(r.ServiceInstallSnapshotName(),req,res)
		if err!=nil{
			r.RemoveConn(addr)
			return err
		}
		return nil
	}
	return errors.New("RPCs.CallInstallSnapshot can not connect to "+addr)
}