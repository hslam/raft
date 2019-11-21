package main

import (
	"hslam.com/git/x/raft"
	"fmt"
	"time"
)

func main() {
	raft.SetLogLevel(raft.DebugLevel)
	node,_:=raft.NewNode("localhost",9001,"node.1",nil,[]*raft.NodeInfo{
		&raft.NodeInfo{Address:"localhost:9001",Data:nil},
		&raft.NodeInfo{Address:"localhost:9002",Data:nil},
		&raft.NodeInfo{Address:"localhost:9003",Data:nil},
	})
	node.Start()
	for{
		fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),node.State(),node.Term(),node.Leader())
		time.Sleep(time.Second*3)
	}
	select {}
}
