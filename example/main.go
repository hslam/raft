package main

import (
	"hslam.com/mgit/Mort/raft"
	"fmt"
	"time"
)

func main() {
	raft.SetLogLevel(raft.TraceLevel)
	go func() {
		server,_:=raft.NewServer(":8001","raft1")
		server.AddNode(":8001")
		server.AddNode(":8002")
		server.AddNode(":8003")	}()
	go func() {
		server,_:=raft.NewServer(":8002","raft2")

		server.AddNode(":8001")
		server.AddNode(":8002")
		server.AddNode(":8003")
	}()

	go func() {
		server,_:=raft.NewServer(":8003","raft3")
		server.ChangeRoleToLeader()
		time.Sleep(time.Second)
		fmt.Println("State:",server.State())
		server.AddNode(":8001")
		server.AddNode(":8002")
		server.AddNode(":8003")
	}()
	select {
	}
}
