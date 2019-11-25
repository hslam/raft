package main

import (
	"hslam.com/git/x/raft"
	"fmt"
	"time"
	"flag"
	"strings"
	"hslam.com/git/x/stats"
)
var(
	host string
	port int
	path string
	join bool
	peers string
	log bool
	benchmark bool
	operation string
	parallel int
	total_calls int
)
func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 9001, "port")
	flag.StringVar(&path, "path", "raft.example/node.1", "data dir")
	flag.BoolVar(&join, "join", false, "")
	flag.StringVar(&peers, "peers", "localhost:9001,false;", "host:port,nonVoting;host:port,nonVoting;")
	flag.BoolVar(&log, "log", true, "log:")
	flag.BoolVar(&benchmark, "b", false, "benchmark")
	flag.StringVar(&operation, "o", "set", "set|lease|readindex")
	flag.IntVar(&parallel, "parallel", 4096, "parallel: -total=10000")
	flag.IntVar(&total_calls, "total", 100000, "total_calls: -total=10000")
}
func main() {
	flag.Parse()
	if log{
		raft.SetLogLevel(raft.InfoLevel)
	}
	var nodes []string
	var infos []*raft.NodeInfo
	if peers != "" {
		nodes = strings.Split(peers, ";")
		for _,v:=range nodes{
			if v=="" {
				continue
			}
			info := strings.Split(v, ",")
			var NonVoting bool
			if len(info)>1{
				if info[1]=="true"{
					NonVoting=true
				}
			}
			infos=append(infos,&raft.NodeInfo{Address:info[0],NonVoting:NonVoting,Data:nil})
		}
	}
	ctx := NewContext()
	node,err:=raft.NewNode(host,port,path,ctx,join,infos)
	if err!=nil{
		panic(err)
	}

	node.RegisterCommand(&Command{})
	node.SetCodec(&raft.JsonCodec{})
	node.Start()
	if benchmark{
		go func() {
			for {
				time.Sleep(time.Second*5)
				if node.IsLeader(){
					node.Do(&Command{Key:"foo",Value:"bar"})
					var Clients =make([]stats.Client,1)
					Clients[0]=&Client{node:node,ctx:ctx,operation:operation}
					stats.StartPrint(parallel,total_calls,Clients)
					break
				}else if node.Leader()!=""{
					break
				}
			}
		}()
	}else {
		for{
			fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),node.State(),node.Term(),node.Leader())
			time.Sleep(time.Second*3)
		}
	}
	select {}
}
