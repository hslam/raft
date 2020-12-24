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
	var closed uint32
	done := make(chan struct{})
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
			node.SetSnapshot(&testSnapshot{ctx: ctx})
			node.SetSyncTypes([]*SyncType{
				{Seconds: 900, Changes: 1},
				{Seconds: 300, Changes: 10},
				{Seconds: 60, Changes: 10000},
			})
			node.Start()
			node.LeaderChange(func() {
				time.Sleep(time.Second)
				leader := node.Leader()
				lookupLeader := node.LookupPeer(leader)
				if lookupLeader != nil && lookupLeader.Address != leader {
					t.Error(lookupLeader.Address, leader)
				}
				node.Do(&testCommand{"foobar"})
				time.Sleep(time.Second)
				if ok := node.Lease(); ok {
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
					if atomic.CompareAndSwapUint32(&closed, 0, 1) {
						close(startJoin)
						close(done)
					}
				}
			})
			if index < 3 {
				<-startJoin
			}
			if node.IsLeader() {
				if _, ok := node.Leave(infos[4].Address); !ok {
					t.Error()
				}
				al.Done()
			} else if index < 3 {
				if _, ok := node.Leave(infos[3].Address); !ok {
					t.Error()
				}
				al.Done()
			}
			al.Wait()
			time.Sleep(time.Second * 3)
			<-done
			node.Stop()
		}()
	}
	wg.Wait()
	os.RemoveAll(dir)
	GetLogLevel()
}