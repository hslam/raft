package main

import (
	"github.com/valyala/fasthttp"
	"hslam.com/mgit/Mort/stats"
	"math/rand"
	"time"
)

func main()  {
	clients:=1
	var wrkClients =make([]stats.Client,clients)
	parallel:=200
	total_calls:=10000
	for i:=0;i<clients;i++{
		var conn =&WrkClient{}
		conn.client=&fasthttp.Client{
			//MaxConnsPerHost:1,
		}
		conn.url="http://localhost:7001/db/"
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

func (c *WrkClient)Call()(int64,bool){
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
		return 0,false
	}
	return int64(length),true
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