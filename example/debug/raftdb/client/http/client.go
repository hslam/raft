package main

import (
	"github.com/valyala/fasthttp"
	"hslam.com/mgit/Mort/stats"
	"math/rand"
	"time"
	"flag"
	"strconv"
)
var port int
var host string
var addr string
var clients int
var total_calls int
var parallel int
func init(){
	flag.StringVar(&host, "h", "127.0.0.1", "host: -h=127.0.0.1")
	flag.IntVar(&port, "p", 7002, "port: -p=7001")
	flag.IntVar(&clients, "clients", 1, "num: -clients=1")
	flag.IntVar(&total_calls, "total", 100000, "total_calls: -total=10000")
	flag.IntVar(&parallel, "parallel", 200, "total_calls: -total=10000")
	flag.Parse()
	addr=host+":"+strconv.Itoa(port)
}
func main()  {
	var wrkClients =make([]stats.Client,clients)
	for i:=0;i<clients;i++{
		var conn =&WrkClient{}
		conn.client=&fasthttp.Client{
			//MaxConnsPerHost:1,
		}
		conn.url="http://"+addr+"/db/"
		conn.meth="POST"
		wrkClients[i]= conn
	}
	stats.StartClientStats(parallel,total_calls,wrkClients)
}

type WrkClient struct {
	client *fasthttp.Client
	url string
	meth string
}

func (c *WrkClient)Call()(int64,int64,bool){
	key:= RandString(4)
	value:= RandString(32)
	var requestBody =[]byte(value)
	req := &fasthttp.Request{}
	req.Header.SetMethod(c.meth)
	if requestBody!=nil{
		req.SetBody(requestBody)
	}
	req.SetRequestURI(c.url+key)
	resp := &fasthttp.Response{}
	err := c.client.Do(req, resp)
	length:=len(resp.Body())
	if err!=nil{
		return int64(len(key)+len(value)),0,false
	}
	return int64(len(key)+len(value)),int64(length),true
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