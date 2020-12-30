// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const minReadIndexLatency = int64(time.Millisecond)

type readIndex struct {
	mu             sync.Mutex
	node           *node
	trigger        chan struct{}
	m              map[uint64][]chan bool
	mutex          sync.Mutex
	latencys       [lastsSize]int64
	latencysCursor int
	min            int64
	done           chan struct{}
	closed         int32
	id             uint64
	working        int32
}

func newReadIndex(n *node) *readIndex {
	r := &readIndex{
		node:    n,
		trigger: make(chan struct{}, 1),
		done:    make(chan struct{}, 1),
		m:       make(map[uint64][]chan bool),
		min:     minReadIndexLatency,
	}
	go r.run()
	return r
}

func (r *readIndex) updateLatency(d int64) (n int64) {
	r.mutex.Lock()
	r.latencysCursor++
	r.latencys[r.latencysCursor%lastsSize] = d
	var min int64 = minReadIndexLatency
	for i := 0; i < lastsSize; i++ {
		if r.latencys[i] > 0 && r.latencys[i] < min {
			min = r.latencys[i]
		}
	}
	//logger.Tracef("readIndex.updateLatency %v,%d", r.latencys, min)
	r.min = min
	r.mutex.Unlock()
	return min
}

func (r *readIndex) minLatency() time.Duration {
	r.mutex.Lock()
	min := r.min
	r.mutex.Unlock()
	//logger.Tracef("readIndex.minLatency%d", min)
	return time.Duration(min)
}

func (r *readIndex) Read() (ok bool) {
	commitIndex := r.node.commitIndex.ID()
	r.node.nodesMut.Lock()
	votingsCount := r.node.votingsCount()
	r.node.nodesMut.Unlock()
	if votingsCount == 1 {
		return true
	}
	var ch = make(chan bool, 1)
	r.mu.Lock()
	if chs, ok := r.m[r.id]; !ok {
		r.m[r.id] = []chan bool{ch}
	} else {
		r.m[r.id] = append(chs, ch)
	}
	r.mu.Unlock()
	select {
	case r.trigger <- struct{}{}:
	default:
	}
	timer := time.NewTimer(defaultCommandTimeout)
	runtime.Gosched()
	select {
	case ok = <-ch:
		if atomic.LoadUint64(&r.node.stateMachine.lastApplied) >= commitIndex {
			timer.Stop()
			return
		}
		var done = make(chan struct{}, 1)
		go r.node.waitApply(commitIndex, done)
		runtime.Gosched()
		select {
		case <-done:
			timer.Stop()
		case <-timer.C:
			ok = false
		}
	case <-timer.C:
		ok = false
	}
	return
}

func (r *readIndex) reply(id uint64, success bool) {
	r.mu.Lock()
	chs, ok := r.m[id]
	delete(r.m, id)
	r.mu.Unlock()
	if ok {
		if len(chs) > 0 {
			for _, ch := range chs {
				ch <- success
			}
		}
	}
}

func (r *readIndex) send() {
	if atomic.CompareAndSwapInt32(&r.working, 0, 1) {
		r.mu.Lock()
		id := r.id
		if chs, ok := r.m[id]; ok {
			if len(chs) > 0 {
				if r.node.isLeader() {
					r.id++
				}
				r.mu.Unlock()
				if !r.node.running {
					r.reply(id, false)
					atomic.StoreInt32(&r.working, 0)
					return
				}
				if r.node.IsLeader() {
					go func(id uint64) {
						start := time.Now().UnixNano()
						ok := r.node.checkLeader()
						go r.updateLatency(time.Now().UnixNano() - start)
						r.reply(id, ok)
					}(id)
					atomic.StoreInt32(&r.working, 0)
					return
				}
				r.reply(id, false)
				atomic.StoreInt32(&r.working, 0)
				return
			}
		}
		r.mu.Unlock()
		atomic.StoreInt32(&r.working, 0)
	}
	return
}

func (r *readIndex) run() {
	for {
	loop:
		runtime.Gosched()
		time.Sleep(r.minLatency() / 10)
		r.send()
		r.mu.Lock()
		length := len(r.m)
		r.mu.Unlock()
		if length > 0 {
			goto loop
		}
		runtime.Gosched()
		select {
		case <-r.trigger:
			goto loop
		case <-r.done:
			return
		}
	}
}

func (r *readIndex) Stop() {
	if !atomic.CompareAndSwapInt32(&r.closed, 0, 1) {
		return
	}
	close(r.done)
	close(r.trigger)
}
