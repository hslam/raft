package main

import (
	"flag"
	"hslam.com/git/x/raft/example/raftdb/node"
	"log"
	"strings"
	_ "net/http/pprof"
	"net/http"
)

var(
	host string
	port int
	rpc_port int
	raft_port int
	addrs string
	data_dir string
)

func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 7001, "port")
	flag.IntVar(&rpc_port, "c", 8001, "port")
	flag.IntVar(&raft_port, "f", 9001, "port")
	flag.StringVar(&addrs, "peers", "localhost:9001,localhost:9002,localhost:9003", "host:port,host:port")
	flag.StringVar(&data_dir, "path", "default.raft.1", "path")
}

func main() {
	go func() {log.Println(http.ListenAndServe(":6061", nil))}()
	flag.Parse()
	var peers []string
	if addrs != "" {
		peers = strings.Split(addrs, ",")
	}
	s := node.NewNode(data_dir, host, port,rpc_port,raft_port,peers)
	log.Fatal(s.ListenAndServe())
}
