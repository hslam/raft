# raft
Package raft implements the Raft distributed consensus protocol.

## Features

* Leader Election
* Log Replication
* Membership Changes
* Log Compaction (Snapshotting)
* [RPC](https://github.com/hslam/rpc "rpc") transport
* [WAL](https://github.com/hslam/wal "wal")
* ReadIndex/Lease Read
* Non-Voting Members (The leader only replicates log entries to them)
* Snapshot Policies (Never/EverySecond/CustomSync/Always)

## Get started

### Install
```
go get github.com/hslam/raft
```
### Import
```
import "github.com/hslam/raft"
```

### [HelloWorld](https://github.com/hslam/raft/tree/master/examples/helloworld "helloworld")

* Leader Election
* Non-Voting Members
* Membership Changes

#### helloworld.go
```go
package main

import (
	"flag"
	"fmt"
	"github.com/hslam/raft"
	"strings"
	"time"
)

var (
	host  string
	port  int
	path  string
	join  bool
	peers string
)

func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 9001, "port")
	flag.StringVar(&path, "path", "raft.helloworld/node.1", "data dir")
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
		for _, v := range nodes {
			if v == "" {
				continue
			}
			info := strings.Split(v, ",")
			var NonVoting bool
			if len(info) > 1 {
				if info[1] == "true" {
					NonVoting = true
				}
			}
			infos = append(infos, &raft.NodeInfo{Address: info[0], NonVoting: NonVoting, Data: nil})
		}
	}
	node, err := raft.NewNode(host, port, path, nil, join, infos)
	if err != nil {
		panic(err)
	}
	node.Start()
	for {
		fmt.Printf("%d State:%s - Leader:%s\n", time.Now().Unix(), node.State(), node.Leader())
		time.Sleep(time.Second * 3)
	}
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

### [Example](https://github.com/hslam/raft/tree/master/examples/example "example")

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
### [RaftDB](https://github.com/hslam/raftdb "raftdb")
The [raftdb](https://github.com/hslam/raftdb  "raftdb") implements a simple distributed key-value datastore, using the raft distributed consensus protocol.

#### Benchmark
Running on three nodes cluster.
##### Write

<img src="https://raw.githubusercontent.com/hslam/raft-benchmark/master/raft-write-qps.png" width = "400" height = "300" alt="write-qps" align=center><img src="https://raw.githubusercontent.com/hslam/raft-benchmark/master/raft-write-p99.png" width = "400" height = "300" alt="write-p99" align=center>

##### Read Index

<img src="https://raw.githubusercontent.com/hslam/raft-benchmark/master/raft-read-qps.png" width = "400" height = "300" alt="read-qps" align=center><img src="https://raw.githubusercontent.com/hslam/raft-benchmark/master/raft-read-p99.png" width = "400" height = "300" alt="read-p99" align=center>

## License
This package is licensed under a MIT license (Copyright (c) 2019 Meng Huang)

## Author
raft was written by Meng Huang.