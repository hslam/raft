package main

import (
	"hslam.com/mgit/Mort/rpc"
	"hslam.com/mgit/Mort/stats"
	"hslam.com/mgit/Mort/raft/example/raftdb/node"
	"math/rand"
	"strconv"
	"runtime"
	"flag"
	"log"
	"fmt"
	"time"
)
var network string
var codec string
var compress string
var host string
var port int
var addr string
var parallel int
var batch bool
var batch_async bool
var multiplexing bool
var clients int
var total_calls int
var bar bool

func init()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&network, "network", "tcp", "network: -network=tcp|ws|fast|http|http2|quic|udp")
	flag.StringVar(&codec, "codec", "pb", "codec: -codec=pb|json|xml")
	flag.StringVar(&compress, "compress", "gzip", "compress: -compress=no|flate|zlib|gzip")
	flag.StringVar(&host, "h", "127.0.0.1", "host: -h=127.0.0.1")
	flag.IntVar(&port, "p", 8001, "port: -p=8001")
	flag.IntVar(&parallel, "parallel", 512, "parallel: -parallel=512")
	flag.IntVar(&total_calls, "total", 500000, "total_calls: -total=10000")
	flag.BoolVar(&batch, "batch", true, "batch: -batch=false")
	flag.BoolVar(&batch_async, "batch_async", true, "batch_async: -batch_async=false")
	flag.BoolVar(&multiplexing, "multiplexing", false, "multiplexing: -multiplexing=false")
	flag.IntVar(&clients, "clients", 8, "num: -clients=1")
	flag.BoolVar(&bar, "bar", false, "bar: -bar=true")
	log.SetFlags(0)
	flag.Parse()
	addr=host+":"+strconv.Itoa(port)
	stats.SetLog(bar)
}

func main()  {
	fmt.Printf("./client -network=%s -codec=%s -compress=%s -h=%s -p=%d -parallel=%d -total=%d -multiplexing=%t -batch=%t -batch_async=%t -clients=%d\n",network,codec,compress,host,port,parallel,total_calls,multiplexing,batch,batch_async,clients)
	var wrkClients []stats.Client
	if clients>1{
		pool,err := rpc.DialsWithMaxRequests(clients,network,addr,codec,parallel)
		if err != nil {
			log.Fatalln("dailing error: ", err)
		}
		pool.SetCompressType(compress)
		if batch {pool.EnableBatch()}
		if batch_async{pool.EnableBatchAsync()}
		if multiplexing{pool.EnableMultiplexing()}
		wrkClients=make([]stats.Client,len(pool.All()))
		for i:=0; i<len(pool.All());i++  {
			wrkClients[i]=&WrkClient{pool.All()[i]}
			conn:= pool.All()[i]
			defer conn.Close()
		}
		parallel=pool.GetMaxRequests()
	}else if clients==1 {
		conn, err:= rpc.DialWithMaxRequests(network,addr,codec,parallel)
		if err != nil {
			log.Fatalln("dailing error: ", err)
		}
		defer conn.Close()
		conn.SetCompressType(compress)
		if batch {conn.EnableBatch()}
		if batch_async{conn.EnableBatchAsync()}
		if multiplexing{conn.EnableMultiplexing()}
		parallel=conn.GetMaxRequests()
		wrkClients=make([]stats.Client,1)
		wrkClients[0]= &WrkClient{conn}
	}else {
		return
	}
	stats.StartClientStats(parallel,total_calls,wrkClients)
}

type WrkClient struct {
	Conn rpc.Client
}

func (c *WrkClient)Call()(int64,int64,bool){
	A:= RandString(4)
	B:= RandString(32)
	req := &node.Request{Key:A,Value:B}
	var res node.Response
	c.Conn.Call("S.Set", req, &res)
	if res.Ok==true{
		return int64(len(A)+len(B)),0,true
	}
	return int64(len(A)+len(B)),0,false
}

func RandString(len int) string {
	r:= rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}