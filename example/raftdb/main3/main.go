package main

import (
	"flag"
	"hslam.com/mgit/Mort/raft/example/raftdb/server"
	"log"
	"os"
	"strings"
)

var host string
var port int
var raft_port int
var join string
var serverPeers string
var path string
func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 4003, "port")
	flag.IntVar(&raft_port, "r", 9003, "port")
	flag.StringVar(&join, "join", "", "host:port of leader to join")
	flag.StringVar(&serverPeers, "peers", ":9001,:9002,:9003", "host:port,host:port")
	flag.StringVar(&path, "path", "raft.3", "path")
}

func main() {

	if err := os.MkdirAll(path, 0744); err != nil {
		log.Fatalf("Unable to create path: %v", err)
	}
	var peers []string
	if serverPeers != "" {
		peers = strings.Split(serverPeers, ",")
	}
	s := server.New(path, host, port,raft_port)
	log.Fatal(s.ListenAndServe(join,peers))
}
