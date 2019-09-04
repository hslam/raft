package main

import (
	"hslam.com/mgit/Mort/raft"
	"fmt"
	"time"
)

func main() {
	raft.SetLogLevel(raft.DebugLevel)
	node,_:=raft.NewNode(":8002","raft2",nil)
	fmt.Println("State:",node.State())
	//node.SetNode([]string{":8001",":8002",":8003",":8004",":8005"})
	node.SetNode([]string{":8001",":8002",":8003"})
	node.Start()

	for{
		time.Sleep(time.Second*3)
		fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),node.State(),node.Term(),node.Leader())
	}
	select {
	}
}
