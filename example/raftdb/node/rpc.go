package node

import (
	"hslam.com/git/x/rpc"
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

func InitRPCProxy(maxConnsPerHost int) {
	for i:=0;i<maxConnsPerHost ;i++  {
		rpcsPool=append(rpcsPool,newRPCs([]string{}) )
	}
	go run()
}
func run(){
	ticker:=time.NewTicker(time.Second)
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
	lastTime 				time.Time
	keepAlive				time.Duration
}
func newRPCs(addrs []string) *RPCs{
	r :=&RPCs{
		conns:make(map[string]*Client),
	}
	for _, addr:= range addrs {
		conn, err := r.NewConn(addr)
		if err==nil{

			c:=&Client{conn,time.Now(),keepAlive}
			r.conns[addr] = c
		}
	}
	return r
}

func (r *RPCs) check(){
	for addr,conn:=range r.conns{
		if conn.lastTime.Add(conn.keepAlive).Before(time.Now()){
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
		c:=&Client{conn,time.Now(),keepAlive}
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
		go func(conn rpc.Client) {
			time.Sleep(keepAlive)
			conn.Close()
		}(conn)
	}
}
func (r *RPCs) NewConn(addr string) (rpc.Client, error){
	opts:=rpc.DefaultOptions()
	opts.SetRetry(false)
	opts.SetCompressType("gzip")
	opts.SetMultiplexing(true)
	opts.SetBatch(true)
	opts.SetBatchAsync(true)
	return rpc.DialWithOptions(network,addr,codec,opts)
}

func (r *RPCs) Call(addr string,req *Request, res *Response)error {
	conn:= r.GetConn(addr)
	if conn!=nil{
		err:=conn.Call("S.Set",req,res)
		if err!=nil{
			r.RemoveConn(addr)
			return err
		}
		conn.lastTime=time.Now()
		return nil
	}
	return errors.New("RPCs.Call can not connect to "+addr)
}
