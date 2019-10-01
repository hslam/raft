package node

import (
	"hslam.com/mgit/Mort/rpc"
	"errors"
	"sync"
)

const (
	network = "tcp"
	codec = "pb"
	MaxConnsPerHost=8
)
var rpcsPool []*RPCs
var rpcsMut sync.Mutex

func init() {
	for i:=0;i<MaxConnsPerHost ;i++  {
		rpcsPool=append(rpcsPool,newRPCs([]string{}) )
	}
}
func getRPCs() (*RPCs) {
	rpcsMut.Lock()
	defer rpcsMut.Unlock()
	if len(rpcsPool) < 1 {
		return nil
	}
	rpcs := rpcsPool[0]
	rpcsPool = append(rpcsPool[1:], rpcs)
	return rpcs
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
			conn.EnableBatch()
			conn.EnableBatchAsync()
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
		conn.EnableBatch()
		conn.EnableBatchAsync()
		r.conns[addr] = conn
		return r.conns[addr]
	}
	return nil
}

func (r *RPCs) RemoveConn(addr string){
	r.mu.Lock()
	defer r.mu.Unlock()
	if conn,ok:=r.conns[addr];ok{
		delete(r.conns,addr)
		defer conn.Close()
	}
}
func (r *RPCs) NewConn(addr string) (rpc.Client, error){
	return rpc.Dial(network,addr,codec)
}

func (r *RPCs) Call(addr string,req *Request, res *Response)error {
	conn:= r.GetConn(addr)
	if conn!=nil{
		err:=conn.Call("S.Set",req,res)
		if err!=nil{
			r.RemoveConn(addr)
			return err
		}
		return nil
	}
	return errors.New("RPCs.Call can not connect to "+addr)
}
