package main

import (
	"hslam.com/git/x/rpc"
	"hslam.com/git/x/stats"
	"hslam.com/git/x/raft/example/raftdb/node"
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
	flag.IntVar(&total_calls, "total", 1000000, "total_calls: -total=10000")
	flag.BoolVar(&batch, "batch", true, "batch: -batch=false")
	flag.BoolVar(&batch_async, "batch_async", true, "batch_async: -batch_async=false")
	flag.BoolVar(&multiplexing, "multiplexing", true, "multiplexing: -multiplexing=false")
	flag.IntVar(&clients, "clients", 8, "num: -clients=1")
	flag.BoolVar(&bar, "bar", false, "bar: -bar=true")
	log.SetFlags(0)
	flag.Parse()
	addr=host+":"+strconv.Itoa(port)
	stats.SetBar(bar)
}

func main()  {
	fmt.Printf("./client -network=%s -codec=%s -compress=%s -h=%s -p=%d -parallel=%d -total=%d -multiplexing=%t -batch=%t -batch_async=%t -clients=%d\n",network,codec,compress,host,port,parallel,total_calls,multiplexing,batch,batch_async,clients)
	var wrkClients []stats.Client
	opts:=rpc.DefaultOptions()
	opts.SetMaxRequests(parallel)
	opts.SetCompressType(compress)
	opts.SetBatch(batch)
	opts.SetBatchAsync(batch_async)
	opts.SetMultiplexing(multiplexing)
	if clients>1{
		pool,err := rpc.DialsWithOptions(clients,network,addr,codec,opts)
		if err != nil {
			log.Fatalln("dailing error: ", err)
		}
		wrkClients=make([]stats.Client,len(pool.All()))
		for i:=0; i<len(pool.All());i++  {
			wrkClients[i]=&WrkClient{pool.All()[i]}
			conn:= pool.All()[i]
			defer conn.Close()
		}
		parallel=pool.GetMaxRequests()
	}else if clients==1 {
		conn, err:= rpc.DialWithOptions(network,addr,codec,opts)
		if err != nil {
			log.Fatalln("dailing error: ", err)
		}
		defer conn.Close()
		parallel=conn.GetMaxRequests()
		wrkClients=make([]stats.Client,1)
		wrkClients[0]= &WrkClient{conn}
	}else {
		return
	}
	stats.StartPrint(parallel,total_calls,wrkClients)
}

type WrkClient struct {
	Conn rpc.Client
}

func (c *WrkClient)Call()(int64,int64,bool){
	A:= "foo"
	req := &node.Request{Key:A}
	var res node.Response
	c.Conn.Call("S.ReadIndexGet", req, &res)
	if res.Ok==true{
		return int64(len(A)),int64(len(res.Result)),true
	}
	return int64(len(A)),int64(len(res.Result)),false
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