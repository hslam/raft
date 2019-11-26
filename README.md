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

## Projects

* [helloworld](https://hslam.com/git/x/raft/src/master/examples/helloworld "helloworld")
* [example](https://hslam.com/git/x/raft/src/master/examples/example "example")
* [raftdb](https://hslam.com/git/x/raftdb "raftdb")


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
./node -h=localhost -p=9001 -path="default.raft/node.1" -join=false -peers="localhost:9001"
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
### [RaftDB](https://hslam.com/git/x/raftdb "raftdb") Benchmark

#### Linux Environment
* **CPU** 12 Cores 3.1 GHz
* **Memory** 24 GiB

```
cluster     operation   transport   requests/s  average(ms) fastest(ms) median(ms)  p99(ms)     slowest(ms)
Singleton   ReadIndex   HTTP        74456       6.63        2.62        6.23        12.12       110.90
Singleton   ReadIndex   RPC         293865      13.14       4.09        12.14       31.22       35.09
Singleton   Write       HTTP        57488       8.79        2.19        7.68        24.00       119.71
Singleton   Write       RPC         132045      30.21       6.39        27.59       70.14       86.11
ThreeNodes  ReadIndex   HTTP        43053       11.72       2.83        7.43        58.92       1125.58
ThreeNodes  ReadIndex   RPC         267685      14.65       4.36        13.47       31.59       44.72
ThreeNodes  Write       HTTP        35241       14.42       4.21        10.41       73.28       114.84
ThreeNodes  Write       RPC         103035      38.82       8.91        38.90       76.74       88.05
```

#### HTTP WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	512
	Parallel calls per client:	1
	Total calls:	100000
	Total time:	2.84s
	Requests per second:	35241.64
	Fastest time for request:	4.21ms
	Average time per request:	14.42ms
	Slowest time for request:	114.84ms

Time:
	10%	time for request:	7.84ms
	50%	time for request:	10.41ms
	90%	time for request:	24.42ms
	99%	time for request:	73.28ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### RPC WRITE THREE NODES BENCHMARK
```
Summary:
	Clients:	8
	Parallel calls per client:	512
	Total calls:	100000
	Total time:	0.97s
	Requests per second:	103035.21
	Fastest time for request:	8.91ms
	Average time per request:	38.82ms
	Slowest time for request:	88.05ms

Time:
	10%	time for request:	21.52ms
	50%	time for request:	38.90ms
	90%	time for request:	54.88ms
	99%	time for request:	76.74ms

Result:
	Response ok:	1000000 (100.00%)
	Errors:	0 (0.00%)
```

#### ETCD WRITE THREE NODES BENCHMARK
```
Summary:
	Conns:	8
	Clients:	512
	Total calls:	100000
	Total time:	2.54s
	Requests per second:	39357.45
	Fastest time for request:	0.90ms
	Average time per request:	12.90ms
	Slowest time for request:	71.50ms

Time:
	10%	time for request:	7.10ms
	50%	time for request:	10.50ms
	90%	time for request:	17.50ms
	99%	time for request:	52.10ms

Result:
	Response ok:	100000 (100.00%)
	Errors:	0 (0.00%)
```

## Licence
This package is licenced under a MIT licence (Copyright (c) 2019 Mort Huang)

## Authors
raft was written by Mort Huang.