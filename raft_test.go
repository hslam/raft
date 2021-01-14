// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testContext struct {
	mut  sync.RWMutex
	data string
}

func (ctx *testContext) Set(value string) {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()
	ctx.data = value
}

func (ctx *testContext) Get() string {
	ctx.mut.RLock()
	defer ctx.mut.RUnlock()
	return ctx.data
}

type testSnapshot struct {
	ctx *testContext
}

func (s *testSnapshot) Save(w io.Writer) (int, error) {
	return w.Write([]byte(s.ctx.Get()))
}

func (s *testSnapshot) Recover(r io.Reader) (int, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	s.ctx.Set(string(raw))
	return len(raw), err
}

type testCommand struct {
	Data string
}

func (c *testCommand) Type() uint64 {
	return 1
}

func (c *testCommand) Do(context interface{}) (interface{}, error) {
	ctx := context.(*testContext)
	ctx.Set(c.Data)
	return nil, nil
}

type testCommand1 struct {
	Data string
}

func (c *testCommand1) Type() uint64 {
	return 2
}

func (c *testCommand1) Do(context interface{}) (interface{}, error) {
	return nil, nil
}

func TestCluster(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}, {Address: "localhost:9004", NonVoting: true}, {Address: "localhost:9005", NonVoting: true}}
	wg := sync.WaitGroup{}
	var readflag uint32
	startRead := make(chan struct{})
	var joinflag uint32
	startJoin := make(chan struct{})
	al := sync.WaitGroup{}
	al.Add(3)
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var n Node
			var err error
			if index < 3 {
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:3])
				if err != nil {
					t.Error(err)
				}
			} else {
				<-startJoin
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, true, infos[:index+1])
				if err != nil {
					t.Error(err)
				}
			}
			node := n.(*node)
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSyncTypes([]*SyncType{
				{Seconds: 1, Changes: 1},
			})
			node.MemberChange(func() {
			})
			node.LeaderChange(func() {
				time.Sleep(time.Second * 2)
				lookupLeader := node.LookupPeer(leader)
				leader := node.Leader()
				if lookupLeader != nil && lookupLeader.Address != leader {
					t.Error(lookupLeader.Address, leader)
				}
				if node.IsLeader() {
					go node.waitApplyTimeout(node.commitIndex.ID()+1, time.NewTimer(time.Millisecond))
					time.Sleep(time.Millisecond * 100)
					go node.waitApplyTimeout(node.commitIndex.ID()+1, time.NewTimer(defaultCommandTimeout))
				}
				node.Do(&testCommand{"foobar"})
				if node.IsLeader() {
					if atomic.CompareAndSwapUint32(&readflag, 0, 1) {
						close(startRead)
					}
				}
				<-startRead
				time.Sleep(time.Second)
				if ok := node.LeaseRead(); ok {
					value := ctx.Get()
					if value != "foobar" {
						t.Error(value)
					}
				}
				if ok := node.ReadIndex(); ok {
					value := ctx.Get()
					if value != "foobar" {
						t.Error(value)
					}
				}
				if node.IsLeader() {
					if atomic.CompareAndSwapUint32(&joinflag, 0, 1) {
						close(startJoin)
					}
				}
			})
			node.Start()
			if index < 3 {
				<-startJoin
				time.Sleep(time.Second * 3)
				if node.IsLeader() {
					if ok := node.Leave(infos[4].Address); !ok {
						t.Error()
					}
					al.Done()
				} else {
					if ok := node.Leave(infos[3].Address); !ok {
						t.Error()
					}
					al.Done()
				}
			}
			al.Wait()
			time.Sleep(time.Second * 3)
			if index < 3 && node.IsLeader() {
				node.nextState()
			}
			time.Sleep(time.Second * 3)
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestClusterMore(t *testing.T) {
	NewNode("localhost", 9001, "", nil, false, nil)
	os.RemoveAll(defaultDataDir)
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}, {Address: "localhost:9004"}, {Address: "localhost:9005"}}
	wg := sync.WaitGroup{}
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var n Node
			var err error
			if index < 3 {
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:3])
				if err != nil {
					t.Error(err)
				}
			} else {
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, true, infos[:index+1])
				if err != nil {
					t.Error(err)
				}
			}
			node := n.(*node)
			if node.Address() != address {
				t.Error()
			}
			node.RegisterCommand(nil)
			node.RegisterCommand(&DefaultCommand{})
			node.registerCommand(nil)
			node.Do(&DefaultCommand{})

			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSnapshotPolicy(Always)
			node.SetSnapshotPolicy(EverySecond)
			node.SetSnapshotPolicy(EveryMinute)
			node.SetSnapshotPolicy(EveryHour)
			node.SetSnapshotPolicy(EveryDay)
			node.SetSnapshotPolicy(DefalutSync)
			node.SetSnapshotPolicy(Never)
			node.SetSyncTypes([]*SyncType{
				{Seconds: 1, Changes: 1},
			})
			node.SetGzipSnapshot(true)
			node.MemberChange(func() {
			})
			var start bool
			node.LeaderChange(func() {
				start = true
			})
			node.Start()
			for {
				time.Sleep(time.Second)
				if start {
					break
				}
			}
			time.Sleep(time.Second * 3)
			if node.isLeader() {
				node.put(nil)
				invoker := node.put(&testCommand1{})
				if invoker.Error != ErrCommandNotRegistered {
					t.Error(invoker.Error)
				}
				node.log.applyCommitedEnd(node.commitIndex.ID())
			}
			node.Stop()
			if index >= 3 {
				node.deleteNotPeers(nil)
			}
			node.log.deleteAfter(node.lastLogIndex)
			node.log.deleteAfter(node.firstLogIndex)
			node.log.deleteAfter(1)
			node.SetLogLevel(node.GetLogLevel())
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestClusterState(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}, {Address: "localhost:9004", NonVoting: true}, {Address: "localhost:9005", NonVoting: true}}
	wg := sync.WaitGroup{}
	al := sync.WaitGroup{}
	al.Add(3)
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var n Node
			var err error
			if index < 3 {
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:3])
				if err != nil {
					t.Error(err)
				}
			} else {
				n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, true, infos[:index+1])
				if err != nil {
					t.Error(err)
				}
			}
			node := n.(*node)
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSyncTypes([]*SyncType{
				{Seconds: 1, Changes: 1},
			})
			node.SetGzipSnapshot(true)
			node.MemberChange(func() {
			})
			node.LeaderChange(func() {
			})
			node.Start()
			time.Sleep(time.Second * 3)
			if index < 3 && node.IsLeader() {
				if !node.Ready() {
					t.Error()
				}
				if len(node.Address()) == 0 {
					t.Error()
				}
				node.stepDown(false)
			}
			time.Sleep(time.Second * 2)
			if index < 3 && node.IsLeader() {
				node.nextState()
			}
			time.Sleep(time.Second * 3)
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestLeaderTimeout(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}}
	wg := sync.WaitGroup{}
	al := sync.WaitGroup{}
	al.Add(3)
	stop := sync.WaitGroup{}
	stop.Add(2)
	count := int32(0)
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var n Node
			var err error
			n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:3])
			if err != nil {
				t.Error(err)
			}
			if index > 0 {
				time.Sleep(time.Second * 5)
			}
			node := n.(*node)
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSyncTypes([]*SyncType{
				{Seconds: 1, Changes: 1},
			})
			node.SetGzipSnapshot(true)
			node.MemberChange(func() {
			})
			var start bool
			node.LeaderChange(func() {
				start = true
			})
			node.Start()
			for {
				time.Sleep(time.Second)
				if start {
					break
				}
			}
			time.Sleep(time.Second * 3)
			if node.IsLeader() {
				stop.Wait()
				time.Sleep(time.Second * 5)
				node.Stop()
			} else {
				time.Sleep(time.Second * 8)
				node.Stop()
				if atomic.AddInt32(&count, 1) < 3 {
					stop.Done()
				}
			}
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestClusterNonVoting(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001", NonVoting: true}, {Address: "localhost:9002", NonVoting: true}, {Address: "localhost:9003"}, {Address: "localhost:9004"}, {Address: "localhost:9005"}}
	wg := sync.WaitGroup{}
	al := sync.WaitGroup{}
	al.Add(3)
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var node Node
			var err error
			node, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos)
			if err != nil {
				t.Error(err)
			}
			if index > 1 {
				time.Sleep(time.Second * 3)
			}
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSnapshotPolicy(Always)
			node.Start()
			time.Sleep(time.Second * 6)
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestSingle(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}}
	var readflag uint32
	startRead := make(chan struct{})
	var joinflag uint32
	startJoin := make(chan struct{})

	address := infos[0].Address
	index := 0
	ctx := &testContext{data: ""}
	strs := strings.Split(address, ":")
	port, _ := strconv.Atoi(strs[1])
	var n Node
	var err error
	n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:1])
	if err != nil {
		t.Error(err)
	}
	node := n.(*node)
	node.RegisterCommand(&testCommand{})
	node.SetCodec(&JSONCodec{})
	node.SetContext(ctx)
	node.SetSnapshot(&testSnapshot{ctx: ctx})
	node.SetSyncTypes([]*SyncType{
		{Seconds: 1, Changes: 1},
	})
	node.LeaderChange(func() {
		time.Sleep(time.Second * 3)
		lookupLeader := node.LookupPeer(leader)
		leader := node.Leader()
		if lookupLeader != nil && lookupLeader.Address != leader {
			t.Error(lookupLeader.Address, leader)
		}
		node.Do(&testCommand{"foobar"})
		if node.IsLeader() {
			if atomic.CompareAndSwapUint32(&readflag, 0, 1) {
				close(startRead)
			}
		}
		<-startRead
		time.Sleep(time.Second * 3)
		if ok := node.LeaseRead(); ok {
			value := ctx.Get()
			if value != "foobar" {
				t.Error(value)
			}
		}
		if ok := node.ReadIndex(); ok {
			value := ctx.Get()
			if value != "foobar" {
				t.Error(value)
			}
		}
		if node.IsLeader() {
			if atomic.CompareAndSwapUint32(&joinflag, 0, 1) {
				close(startJoin)
			}
		}
	})
	node.Start()
	<-startJoin
	time.Sleep(time.Second * 3)
	if index < 3 && node.IsLeader() {
		node.nextState()
	}
	time.Sleep(time.Second * 3)
	node.Stop()
	os.RemoveAll(dir)
}

func TestStateMachine(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}}
	{
		var readflag uint32
		startRead := make(chan struct{})
		var joinflag uint32
		startJoin := make(chan struct{})

		address := infos[0].Address
		index := 0
		ctx := &testContext{data: ""}
		strs := strings.Split(address, ":")
		port, _ := strconv.Atoi(strs[1])
		var n Node
		var err error
		n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:1])
		if err != nil {
			t.Error(err)
		}
		node := n.(*node)
		node.RegisterCommand(&testCommand{})
		node.SetCodec(&JSONCodec{})
		node.SetContext(ctx)
		node.SetSnapshot(&testSnapshot{ctx: ctx})
		node.SetSyncTypes([]*SyncType{
			{Seconds: 1, Changes: 1},
		})
		node.LeaderChange(func() {
			time.Sleep(time.Second * 3)
			lookupLeader := node.LookupPeer(leader)
			leader := node.Leader()
			if lookupLeader != nil && lookupLeader.Address != leader {
				t.Error(lookupLeader.Address, leader)
			}
			node.Do(&testCommand{"foobar"})
			if node.IsLeader() {
				if atomic.CompareAndSwapUint32(&readflag, 0, 1) {
					close(startRead)
				}
			}
			<-startRead
			time.Sleep(time.Second * 3)
			if ok := node.LeaseRead(); ok {
				value := ctx.Get()
				if value != "foobar" {
					t.Error(value)
				}
			}
			if ok := node.ReadIndex(); ok {
				value := ctx.Get()
				if value != "foobar" {
					t.Error(value)
				}
			}
			if node.IsLeader() {
				if atomic.CompareAndSwapUint32(&joinflag, 0, 1) {
					close(startJoin)
				}
			}
		})
		node.Start()
		<-startJoin
		time.Sleep(time.Second * 3)
		node.Stop()
		node.log.wal.Reset()
	}
	os.Remove(dir + "/node.1/commitindex")
	time.Sleep(time.Second * 3)
	{
		var readflag uint32
		startRead := make(chan struct{})
		var joinflag uint32
		startJoin := make(chan struct{})

		address := infos[0].Address
		index := 0
		ctx := &testContext{data: ""}
		strs := strings.Split(address, ":")
		port, _ := strconv.Atoi(strs[1])
		var n Node
		var err error
		n, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:1])
		if err != nil {
			t.Error(err)
		}
		node := n.(*node)
		node.RegisterCommand(&testCommand{})
		node.SetCodec(&JSONCodec{})
		node.SetContext(ctx)
		node.SetSnapshot(&testSnapshot{ctx: ctx})
		node.SetSyncTypes([]*SyncType{
			{Seconds: 1, Changes: 1},
		})
		node.LeaderChange(func() {
			time.Sleep(time.Second * 3)
			lookupLeader := node.LookupPeer(leader)
			leader := node.Leader()
			if lookupLeader != nil && lookupLeader.Address != leader {
				t.Error(lookupLeader.Address, leader)
			}
			node.Do(&testCommand{"foobar"})
			if node.IsLeader() {
				if atomic.CompareAndSwapUint32(&readflag, 0, 1) {
					close(startRead)
				}
			}
			<-startRead
			time.Sleep(time.Second * 3)
			if ok := node.LeaseRead(); ok {
				value := ctx.Get()
				if value != "foobar" {
					t.Error(value)
				}
			}
			if ok := node.ReadIndex(); ok {
				value := ctx.Get()
				if value != "foobar" {
					t.Error(value)
				}
			}
			if node.IsLeader() {
				if atomic.CompareAndSwapUint32(&joinflag, 0, 1) {
					close(startJoin)
				}
			}
		})
		node.Start()
		<-startJoin
		time.Sleep(time.Second * 3)
		node.Stop()
	}
	os.RemoveAll(dir)
}

func TestClusterMeta(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}}
	wg := sync.WaitGroup{}
	for i := 0; i < len(infos); i++ {
		address := infos[i].Address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &testContext{data: ""}
			strs := strings.Split(address, ":")
			port, _ := strconv.Atoi(strs[1])
			var node Node
			var err error
			node, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos)
			if err != nil {
				t.Error(err)
			}
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSnapshotPolicy(Always)
			node.Start()
			time.Sleep(time.Second * 5)
			for j := 0; j < len(infos); j++ {
				addr := infos[j].Address
				ok := node.SetNodeMeta(addr, []byte(addr))
				if !ok {
					t.Error()
				}
				meta, ok := node.GetNodeMeta(addr)
				if !ok {
					t.Error()
				} else if string(meta) != addr {
					t.Error()
				}
			}
			time.Sleep(time.Second * 5)
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}
