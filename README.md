# raft
Package raft implements the Raft distributed consensus protocol.

## Features

* Leader Election
* Log Replication
* Membership Changes
* Log Compaction (Snapshotting)
* [RPC](https://github.com/hslam/rpc "rpc") transport
* [WAL](https://github.com/hslam/wal "wal")
* ReadIndex/LeaseRead
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

### Example

```go
package main

import (
	"flag"
	"github.com/hslam/raft"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
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
	flag.StringVar(&path, "path", "raft.example/node.1", "data dir")
	flag.BoolVar(&join, "join", false, "")
	flag.StringVar(&peers, "peers", "localhost:9001", "host:port,nonVoting;host:port")
	flag.Parse()
}

func main() {
	ctx := NewContext()
	node, err := raft.NewNode(host, port, path, ctx, join, parse(peers))
	if err != nil {
		panic(err)
	}
	node.RegisterCommand(&Command{})
	node.SetCodec(&raft.JSONCodec{})
	node.SetSnapshot(NewSnapshot(ctx))
	node.SetSyncTypes([]*raft.SyncType{
		{Seconds: 900, Changes: 1},
		{Seconds: 300, Changes: 10},
		{Seconds: 60, Changes: 10000},
	})
	node.Start()
	node.LeaderChange(func() {
		if node.IsLeader() {
			node.Do(&Command{"foobar"})
			log.Printf("State:%s, Set:foobar\n", node.State())
			if ok := node.ReadIndex(); ok {
				log.Printf("State:%s, Get:%s\n", node.State(), ctx.Get())
			}
		} else {
			time.Sleep(time.Second * 3)
			log.Printf("State:%s, Get:%s\n", node.State(), ctx.Get())
		}
	})
	for {
		time.Sleep(time.Second * 5)
		log.Printf("State:%s, Leader:%s\n", node.State(), node.Leader())
	}
}

// Command implements the raft.Command interface.
type Command struct {
	Data string
}

func (c *Command) Type() int32 {
	return 1
}

func (c *Command) Do(context interface{}) (interface{}, error) {
	ctx := context.(*Context)
	ctx.Set(c.Data)
	return nil, nil
}

type Context struct {
	mut  sync.RWMutex
	data string
}

func NewContext() *Context {
	return &Context{data: ""}
}

func (ctx *Context) Set(value string) {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()
	ctx.data = value
}

func (ctx *Context) Get() string {
	ctx.mut.RLock()
	defer ctx.mut.RUnlock()
	return ctx.data
}

// Snapshot implements the raft.Snapshot interface.
type Snapshot struct {
	context *Context
}

func NewSnapshot(context *Context) *Snapshot {
	return &Snapshot{context: context}
}

func (s *Snapshot) Save(w io.Writer) (int, error) {
	return w.Write([]byte(s.context.Get()))
}

func (s *Snapshot) Recover(r io.Reader) (int, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	s.context.Set(string(raw))
	return len(raw), err
}

func parse(peers string) (infos []*raft.NodeInfo) {
	if peers != "" {
		nodes := strings.Split(peers, ";")
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
			nodeInfo := &raft.NodeInfo{
				Address:   info[0],
				NonVoting: NonVoting,
				Data:      nil,
			}
			infos = append(infos, nodeInfo)
		}
	}
	return
}
```

### Build
```
go build -o node main.go
```

**one node**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.1" -join=false \
-peers="localhost:9001"
```

**three nodes**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.hw.1" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"

./node -h=localhost -p=9002 -path="raft.example/node.hw.2" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"

./node -h=localhost -p=9003 -path="raft.example/node.hw.3" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"
```

**non-voting**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.nv.1" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"

./node -h=localhost -p=9002 -path="raft.example/node.nv.2" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"

./node -h=localhost -p=9003 -path="raft.example/node.nv.3" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"

./node -h=localhost -p=9004 -path="raft.example/node.nv.4" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003;localhost:9004,true"
```

**membership changes**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.mc.1" -join=false

./node -h=localhost -p=9002 -path="raft.example/node.mc.2" -join=true \
-peers="localhost:9001;localhost:9002"

./node -h=localhost -p=9003 -path="raft.example/node.mc.3" -join=true \
-peers="localhost:9001;localhost:9002;localhost:9003"
```

## License
This package is licensed under a MIT license (Copyright (c) 2019 Meng Huang)

## Author
raft was written by Meng Huang.