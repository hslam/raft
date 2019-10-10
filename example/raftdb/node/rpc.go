package node

import (
	"hslam.com/mgit/Mort/rpc"
	"errors"
	"sync"
	"time"
)

const (
	network = "tcp"
	codec = "pb"
	MaxConnsPerHost=8
	keepAlive=time.Minute
)
var rpcsPool []*RPCs
var rpcsMut sync.Mutex

func init() {
	for i:=0;i<MaxConnsPerHost ;i++  {
		rpcsPool=append(rpcsPool,newRPCs([]string{}) )
	}
	go run()
}
func run(){
	ticker:=time.NewTicker(keepAlive/2)
	for range ticker.C {
		for _,r:=range rpcsPool{
			r.check()
		}
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
	conns			map[string]*Client
}
type Client struct {
	rpc.Client
	startTime 				time.Time
	keepAlive				time.Duration
}
func newRPCs(addrs []string) *RPCs{
	r :=&RPCs{
		conns:make(map[string]*Client),
	}
	for _, addr:= range addrs {
		conn, err := r.NewConn(addr)
		if err==nil{
			conn.DisableRetry()
			conn.SetCompressType("gzip")
			conn.EnableBatch()
			conn.EnableBatchAsync()
			c:=&Client{conn,time.Now(),keepAlive}
			r.conns[addr] = c
		}
	}
	return r
}

func (r *RPCs) check(){
	for addr,conn:=range r.conns{
		if conn.startTime.Add(conn.keepAlive).Before(time.Now()){
			r.RemoveConn(addr)
		}
	}
}
func (r *RPCs) GetConn(addr string) *Client {
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
		c:=&Client{conn,time.Now(),keepAlive}
		r.conns[addr] = c
		r.conns[addr] = c
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
		conn.startTime=time.Now()
		return nil
	}
	return errors.New("RPCs.Call can not connect to "+addr)
}
