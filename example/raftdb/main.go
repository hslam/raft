package main

import (
	"flag"
	"log"
	"strings"
	"runtime"
	"hslam.com/mgit/Mort/raft/example/raftdb/node"
)

var(
	host string
	port int
	rpc_port int
	raft_port int
	addrs string
	data_dir string
	max int
)

func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 7001, "port")
	flag.IntVar(&rpc_port, "c", 8001, "port")
	flag.IntVar(&raft_port, "f", 9001, "port")
	flag.StringVar(&addrs, "peers", "", "host:port,host:port")
	flag.StringVar(&data_dir, "path", "raft.1", "path")
	flag.IntVar(&max, "m", 8, "MaxConnsPerHost: -m=8")
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	var peers []string
	if addrs != "" {
		peers = strings.Split(addrs, ",")
	}
	s := node.NewNode(data_dir, host, port,rpc_port,raft_port,peers)
	node.InitRPCProxy(max)
	log.Fatal(s.ListenAndServe())
}
