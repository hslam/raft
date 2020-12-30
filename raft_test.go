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

func (c *testCommand) Type() int32 {
	return 1
}

func (c *testCommand) Do(context interface{}) (interface{}, error) {
	ctx := context.(*testContext)
	ctx.Set(c.Data)
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
			var node Node
			var err error
			if index < 3 {
				node, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, false, infos[:3])
				if err != nil {
					t.Error(err)
				}
			} else {
				<-startJoin
				node, err = NewNode(strs[0], port, dir+"/node."+strconv.FormatInt(int64(index), 10), ctx, true, infos[:index+1])
				if err != nil {
					t.Error(err)
				}
			}
			node.RegisterCommand(&testCommand{})
			node.SetCodec(&JSONCodec{})
			node.SetContext(ctx)
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSyncTypes([]*SyncType{
				{Seconds: 1, Changes: 1},
			})
			node.LeaderChange(func() {
				time.Sleep(time.Second * 2)
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
				time.Sleep(time.Second)
				if node.IsLeader() {
					if _, ok := node.Leave(infos[4].Address); !ok {
						t.Error()
					}
					al.Done()
				} else {
					if _, ok := node.Leave(infos[3].Address); !ok {
						t.Error()
					}
					al.Done()
				}
			}
			al.Wait()
			time.Sleep(time.Second * 2)
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
}

func TestClusterMore(t *testing.T) {
	dir := "raft.test"
	os.RemoveAll(dir)
	infos := []*NodeInfo{{Address: "localhost:9001"}, {Address: "localhost:9002"}, {Address: "localhost:9003"}, {Address: "localhost:9004"}, {Address: "localhost:9005"}}
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
			node.SetSnapshotPolicy(Always)
			node.SetSnapshotPolicy(EverySecond)
			node.SetSnapshotPolicy(EveryMinute)
			node.SetSnapshotPolicy(EveryHour)
			node.SetSnapshotPolicy(EveryDay)
			node.SetSnapshotPolicy(DefalutSync)
			node.SetSnapshotPolicy(Never)
			node.ClearSyncType()
			node.AppendSyncType(1, 1)
			node.SetGzipSnapshot(true)
			node.LeaderChange(func() {
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
				time.Sleep(time.Second)
				if node.IsLeader() {
					if _, ok := node.Leave(infos[4].Address); !ok {
						t.Error()
					}
					al.Done()
				} else {
					if _, ok := node.Leave(infos[3].Address); !ok {
						t.Error()
					}
					al.Done()
				}
			}
			al.Wait()
			if index < 3 && node.IsLeader() {
				if !node.Running() {
					t.Error()
				}
				if !node.Ready() {
					t.Error()
				}
				if node.Term() == 0 {
					t.Error()
				}
				if len(node.Address()) == 0 {
					t.Error()
				}
				if _, ok := node.Context().(*testContext); !ok {
					t.Error()
				}
				node.stepDown()
			}
			time.Sleep(time.Second * 3)
			if index < 3 && node.IsLeader() {
				node.nextState()
			}
			time.Sleep(time.Second * 2)
			node.Stop()
			if !node.Stoped() {
				t.Error()
			}
			if index >= 3 {
				node.deleteNotPeers(nil)
			}
			node.log.deleteAfter(node.lastLogIndex)
			node.log.deleteAfter(node.firstLogIndex)
			node.log.deleteAfter(1)
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
	GetLogLevel()
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
