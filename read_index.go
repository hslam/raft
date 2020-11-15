// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

type readIndex struct {
	mu      sync.Mutex
	node    *node
	trigger chan struct{}
	m       map[uint64][]chan bool
	done    chan struct{}
	closed  int32
	id      uint64
	working int32
}

func newReadIndex(n *node) *readIndex {
	r := &readIndex{
		node:    n,
		trigger: make(chan struct{}, 1),
		done:    make(chan struct{}, 1),
		m:       make(map[uint64][]chan bool),
	}
	go r.run()
	return r
}

func (r *readIndex) Read() (ok bool) {
	commitIndex := atomic.LoadUint64(&r.node.commitIndex)
	r.node.nodesMut.Lock()
	votingsCount := r.node.votingsCount()
	r.node.nodesMut.Unlock()
	if votingsCount == 1 {
		return true
	}
	var ch = make(chan bool, 1)
	var applied bool
	var done chan struct{}
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
	if atomic.LoadUint64(&r.node.stateMachine.lastApplied) >= commitIndex {
		applied = true
	}
	if !applied {
		done = make(chan struct{}, 1)
		go r.waitApply(commitIndex, done)
	}
	var timeout bool
	select {
	case ok = <-ch:
	case <-timer.C:
		ok = false
		timeout = true
	}
	if !timeout {
		if !applied {
			select {
			case <-done:
				timer.Stop()
			case <-timer.C:
				ok = false
			}
		} else {
			timer.Stop()
		}
	}
	return
}

func (r *readIndex) waitApply(commitIndex uint64, done chan struct{}) {
	for {
		if atomic.LoadUint64(&r.node.stateMachine.lastApplied) >= commitIndex {
			select {
			case done <- struct{}{}:
			default:
			}
			return
		}
		time.Sleep(time.Duration(minLatency) / 10)
	}
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
		defer atomic.StoreInt32(&r.working, 0)
		defer func() {
			if err := recover(); err != nil {
			}
		}()
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
					return
				}
				if r.node.IsLeader() {
					go func(id uint64) {
						r.reply(id, r.node.checkLeader())
					}(id)
					return
				}
				r.reply(id, false)
				return
			}
		}
		r.mu.Unlock()
	}
	return
}

func (r *readIndex) run() {
	for {
	loop:
		time.Sleep(time.Duration(minLatency) / 10)
		r.send()
		r.mu.Lock()
		length := len(r.m)
		r.mu.Unlock()
		if length > 0 {
			goto loop
		}
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
