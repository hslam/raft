# raft
Package raft implements the Raft distributed consensus protocol based on [hslam/rpc](https://github.com/hslam/rpc "rpc").

## Features

* Leader election
* Log replication
* Membership changes
* Log compaction (snapshotting)
* [RPC](https://github.com/hslam/rpc "rpc") transport
* [Write-ahead logging](https://github.com/hslam/wal "wal")
* ReadIndex/LeaseRead
* Non-voting members (The leader only replicates log entries to them)
* Snapshot policies (Never/EverySecond/CustomSync/Always)

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
	"time"
)

var host, path, peers string
var port int
var join bool

func init() {
	flag.StringVar(&host, "h", "localhost", "hostname")
	flag.IntVar(&port, "p", 9001, "port")
	flag.StringVar(&path, "path", "raft.example/node.1", "data dir")
	flag.BoolVar(&join, "join", false, "")
	flag.StringVar(&peers, "peers", "localhost:9001", "host:port,nonVoting;host:port")
	flag.Parse()
}

func main() {
	ctx := &Context{data: ""}
	node, err := raft.NewNode(host, port, path, ctx, join, parse(peers))
	if err != nil {
		panic(err)
	}
	node.RegisterCommand(&Command{})
	node.SetCodec(&raft.JSONCodec{})
	node.SetSnapshot(ctx)
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
	node.Start()
	for range time.NewTicker(time.Second * 5).C {
		log.Printf("State:%s, Leader:%s\n", node.State(), node.Leader())
	}
}

type Context struct{ data string }

func (ctx *Context) Set(value string) { ctx.data = value }
func (ctx *Context) Get() string      { return ctx.data }

// Save implements the raft.Snapshot Save method.
func (ctx *Context) Save(w io.Writer) (int, error) { return w.Write([]byte(ctx.Get())) }

// Recover implements the raft.Snapshot Recover method.
func (ctx *Context) Recover(r io.Reader) (int, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	ctx.Set(string(raw))
	return len(raw), nil
}

// Command implements the raft.Command interface.
type Command struct{ Data string }

func (c *Command) Type() uint64 { return 1 }
func (c *Command) Do(context interface{}) (interface{}, error) {
	context.(*Context).Set(c.Data)
	return nil, nil
}

func parse(peers string) (infos []*raft.NodeInfo) {
	if peers != "" {
		for _, v := range strings.Split(peers, ";") {
			if len(v) > 0 {
				info := strings.Split(v, ",")
				infos = append(infos, &raft.NodeInfo{Address: info[0]})
				if len(info) > 1 && info[1] == "true" {
					infos[len(infos)-1].NonVoting = true
				}
			}
		}
	}
	return
}
```

### Build
```
go build -o node main.go
```

**One node**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.1" -join=false
```

**Three nodes**
```sh
./node -h=localhost -p=9001 -path="raft.example/node.hw.1" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"

./node -h=localhost -p=9002 -path="raft.example/node.hw.2" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"

./node -h=localhost -p=9003 -path="raft.example/node.hw.3" -join=false \
-peers="localhost:9001;localhost:9002;localhost:9003"
```

**Non-voting**
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

**Membership changes**
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