# RAFT
A Golang implementation of the Raft distributed consensus protocol.

## Features

* Leader Election
* Log Replication
* Membership Changes
* Log Compaction (Snapshotting)
* [RPC](https://hslam.com/git/x/rpc "rpc") transport
* ReadIndex/Lease Read
* Non-Voting Members (The leader only replicates log entries to them)
* Snapshot Policies (Never/EverySecond/DefalutSync/CustomSync/Always)

## Get started

### Install
```
go get hslam.com/git/x/raft
```
### Import
```
import "hslam.com/git/x/raft"
```

## Projects

* [HelloWorld](https://hslam.com/git/x/raft/src/master/examples/helloworld "helloworld")
* [Example](https://hslam.com/git/x/raft/src/master/examples/example "example")
* [RaftDB](https://hslam.com/git/x/raftdb "raftdb")

### [HelloWorld](https://hslam.com/git/x/raft/src/master/examples/helloworld "helloworld")

* Leader Election
* Non-Voting Members
* Membership Changes

#### helloworld.go
```
package main

import (
	"hslam.com/git/x/raft"
	"fmt"
	"time"
	"flag"
	"strings"
)
var(
	host string
	port int
	path string
	join bool
	peers string
)
func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 9001, "port")
	flag.StringVar(&path, "path", "node.1", "data dir")
	flag.BoolVar(&join, "join", false, "")
	flag.StringVar(&peers, "peers", "localhost:9001,false;", "host:port,nonVoting;host:port,nonVoting;")
}
func main() {
	flag.Parse()
	raft.SetLogLevel(raft.DebugLevel)
	var nodes []string
	var infos []*raft.NodeInfo
	if peers != "" {
		nodes = strings.Split(peers, ";")
		for _,v:=range nodes{
			info := strings.Split(v, ",")
			var NonVoting bool
			if len(info)>1{
				if info[1]=="true"{
					NonVoting=true
				}
			}
			infos=append(infos,&raft.NodeInfo{Address:info[0],NonVoting:NonVoting,Data:nil})
		}
	}
	node,err:=raft.NewNode(host,port,path,nil,join,infos)
	if err!=nil{
		panic(err)
	}
	node.Start()
	for{
		fmt.Printf("%d State:%s - Term:%d - Leader:%s\n",time.Now().Unix(),node.State(),node.Term(),node.Leader())
		time.Sleep(time.Second*3)
	}
	select {}
}
```
**one node**
```sh
./node -h=localhost -p=9001 -path="raft.helloworld/node.1" -join=false -peers="localhost:9001"
```
**three nodes**
```sh
./node -h=localhost -p=9001 -path="raft.helloworld/node.hw.1" -join=false -peers="localhost:9001;localhost:9002;localhost:9003"
./node -h=localhost -p=9002 -path="raft.helloworld/node.hw.2" -join=false -peers="localhost:9001;localhost:9002;localhost:9003"
./node -h=localhost -p=9003 -path="raft.helloworld/node.hw.3" -join=false -peers="localhost:9001;localhost:9002;localhost:9003"
```

**non-voting**
```sh
./node -h=localhost -p=9001 -path="raft.helloworld/node.nv.1" -join=false -peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"
./node -h=localhost -p=9002 -path="raft.helloworld/node.nv.2" -join=false -peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"
./node -h=localhost -p=9003 -path="raft.helloworld/node.nv.3" -join=false -peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"
./node -h=localhost -p=9004 -path="raft.helloworld/node.nv.4" -join=false -peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"
```

**membership changes**
```sh
./node -h=localhost -p=9001 -path="raft.helloworld/node.mc.1" -join=false
./node -h=localhost -p=9002 -path="raft.helloworld/node.mc.2" -join=true -peers="localhost:9001;localhost:9002"
./node -h=localhost -p=9003 -path="raft.helloworld/node.mc.3" -join=true -peers="localhost:9001;localhost:9002;localhost:9003"
```

### [Example](https://hslam.com/git/x/raft/src/master/examples/example "example")

* Leader Election
* Non-Voting Members
* Membership Changes
* Log Replication
* Log Compaction (Snapshotting)
* ReadIndex/Lease Read
* Snapshot Policies
* Benchmark

**one node set**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.1" -join=false -peers="localhost:9001" -log=true -b=true  -o=set -parallel=4096 -total=100000
```
**one node readindex**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.1" -join=false -peers="localhost:9001" -log=true -b=true  -o=readindex -parallel=4096 -total=100000
```
**one node lease**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.1" -join=false -peers="localhost:9001" -log=true -b=true  -o=lease -parallel=4096 -total=100000
```
**three nodes set**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.tn.1" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=set -parallel=4096 -total=100000
./example -h=localhost -p=9002 -path="raft.example/node.tn.2" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=set -parallel=4096 -total=100000
./example -h=localhost -p=9003 -path="raft.example/node.tn.3" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=set -parallel=4096 -total=100000
```
**three nodes readindex**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.tn.1" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=readindex -parallel=4096 -total=100000
./example -h=localhost -p=9002 -path="raft.example/node.tn.2" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=readindex -parallel=4096 -total=100000
./example -h=localhost -p=9003 -path="raft.example/node.tn.3" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=readindex -parallel=4096 -total=100000
```
**three nodes lease**
```sh
./example -h=localhost -p=9001 -path="raft.example/node.tn.1" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=lease -parallel=4096 -total=100000
./example -h=localhost -p=9002 -path="raft.example/node.tn.2" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=lease -parallel=4096 -total=100000
./example -h=localhost -p=9003 -path="raft.example/node.tn.3" -join=false -peers="localhost:9001;localhost:9002;localhost:9003" -log=true -b=true    -o=lease -parallel=4096 -total=100000
```
### [RaftDB](https://hslam.com/git/x/raftdb "raftdb")
The [raftdb](https://hslam.com/git/x/raftdb  "raftdb") is an example usage of raft library.

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

**Benchmark**
```
cluster    operation transport requests/s average fastest median  p99      slowest
Singleton  ReadIndex HTTP      73348      6.77ms  2.59ms  6.29ms  14.23ms  116.32ms
Singleton  Write     HTTP      60671      8.27ms  2.47ms  7.34ms  23.08ms  134.83ms
ThreeNodes ReadIndex HTTP      54642      9.31ms  2.97ms  7.85ms  69.98ms  106.80ms
ThreeNodes Write     HTTP      37647      13.39ms 4.62ms  9.61ms  77.57ms  142.90ms
Singleton  ReadIndex RPC       310222     12.51ms 4.36ms  12.07ms 29.79ms  34.28ms
Singleton  Write     RPC       138411     28.92ms 6.09ms  24.64ms 103.57ms 121.66ms
ThreeNodes ReadIndex RPC       285650     13.40ms 4.27ms  12.49ms 29.01ms  32.91ms
ThreeNodes Write     RPC       118325     33.74ms 9.76ms  33.40ms 71.38ms  81.32ms
```
## Licence
This package is licenced under a MIT licence (Copyright (c) 2019 Mort Huang)

## Authors
raft was written by Mort Huang.