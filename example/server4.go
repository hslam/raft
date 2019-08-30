package main

import (
	"hslam.com/mgit/Mort/raft"
	"fmt"
	"time"
)

func main() {
	raft.SetLogLevel(raft.DebugLevel)
	server,_:=raft.NewServer(":8004","raft4")
	fmt.Println("State:",server.State())
	server.AddNode(":8001")
	server.AddNode(":8002")
	server.AddNode(":8003")
	server.AddNode(":8004")
	server.AddNode(":8005")
	for{
		time.Sleep(time.Second*3)
		fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),server.State(),server.Term(),server.Leader())
	}
	select {
	}
}
