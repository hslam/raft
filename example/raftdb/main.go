package main

import (
	"flag"
	"log"
	"strings"
	"runtime"
	"strconv"
	_ "net/http/pprof"
	"net/http"
	"hslam.com/git/x/raft/example/raftdb/node"
)

var(
	host string
	port int
	rpc_port int
	raft_port int
	debug bool
	debug_port int
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
	flag.BoolVar(&debug, "debug", true, "debug: -debug=false")
	flag.IntVar(&debug_port, "d", 6061, "debug_port: -dp=6060")
	flag.StringVar(&data_dir, "path", "raft.1", "path")
	flag.IntVar(&max, "m", 8, "MaxConnsPerHost: -m=8")
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	go func() {if debug{log.Println(http.ListenAndServe(":"+strconv.Itoa(debug_port), nil))}}()
	var peers []string
	if addrs != "" {
		peers = strings.Split(addrs, ",")
	}
	s := node.NewNode(data_dir, host, port,rpc_port,raft_port,peers)
	node.InitRPCProxy(max)
	log.Fatal(s.ListenAndServe())
}
