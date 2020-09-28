package main

import (
	"flag"
	"fmt"
	"github.com/hslam/raft"
	"strings"
	"time"
)

var (
	host  string
	port  int
	path  string
	join  bool
	peers string
)

func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 9001, "port")
	flag.StringVar(&path, "path", "raft.helloworld/node.1", "data dir")
	flag.BoolVar(&join, "join", false, "")
	flag.StringVar(&peers, "peers", "localhost:9001,false;", "host:port,nonVoting;host:port,nonVoting;")
}
func main() {
	flag.Parse()
	raft.SetLogLevel(raft.DebugLevel)
	var nodes []string
	var infos []*raft.NodeInfo
	if peers != "" {
		nodes = strings.Split(peers, ";")
		for _, v := range nodes {
			if v == "" {
				continue
			}
			info := strings.Split(v, ",")
			var NonVoting bool
			if len(info) > 1 {
				if info[1] == "true" {
					NonVoting = true
				}
			}
			infos = append(infos, &raft.NodeInfo{Address: info[0], NonVoting: NonVoting, Data: nil})
		}
	}
	node, err := raft.NewNode(host, port, path, nil, join, infos)
	if err != nil {
		panic(err)
	}
	node.Start()
	for {
		fmt.Printf("%d State:%s - Leader:%s\n", time.Now().Unix(), node.State(), node.Leader())
		time.Sleep(time.Second * 3)
	}
	select {}
}
