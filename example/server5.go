package main

import (
	"hslam.com/mgit/Mort/raft"
	"fmt"
	"time"
)

func main() {
	raft.SetLogLevel(raft.DebugLevel)
	node,_:=raft.NewNode(":8005","raft5")
	fmt.Println("State:",node.State())
	node.AddNode(":8001")
	node.AddNode(":8002")
	node.AddNode(":8003")
	node.AddNode(":8004")
	node.AddNode(":8005")
	node.Start()
	for{
		time.Sleep(time.Second*3)
		fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),node.State(),node.Term(),node.Leader())
	}
	select {
	}
}
