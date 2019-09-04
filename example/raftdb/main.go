package main

import (
	"flag"
	"hslam.com/mgit/Mort/raft/example/raftdb/server"
	"log"
	"os"
)

var host string
var port int
var raft_port int
var join string
var path string
func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 4001, "port")
	flag.IntVar(&raft_port, "r", 9001, "port")
	flag.StringVar(&join, "join", "", "host:port of leader to join")
	flag.StringVar(&path, "path", "raft.1", "path")

}

func main() {

	if err := os.MkdirAll(path, 0744); err != nil {
		log.Fatalf("Unable to create path: %v", err)
	}
	s := server.New(path, host, port,raft_port)
	log.Fatal(s.ListenAndServe(join))
}
