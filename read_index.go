// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

type readIndex struct {
	mu       sync.Mutex
	node     *node
	readChan chan chan bool
	m        map[uint64][]chan bool
	id       uint64
	working  int32
}

func newReadIndex(n *node) *readIndex {
	r := &readIndex{
		node:     n,
		readChan: make(chan chan bool, defaultMaxConcurrencyRead),
		m:        make(map[uint64][]chan bool),
	}
	go r.run()
	return r
}

func (r *readIndex) Read() (ok bool) {
	var ch = make(chan bool, 1)
	r.readChan <- ch
	timer := time.NewTimer(defaultCommandTimeout)
	select {
	case ok = <-ch:
		timer.Stop()
	case <-timer.C:
		ok = false
	}
	return
}

func (r *readIndex) reply(id uint64, success bool) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[id]; ok {
		if len(r.m[id]) > 0 {
			for _, ch := range r.m[id] {
				ch <- success
			}
		}
	}
	delete(r.m, id)
}
func (r *readIndex) Update() bool {
	if atomic.CompareAndSwapInt32(&r.working, 0, 1) {
		defer atomic.StoreInt32(&r.working, 0)
		defer func() {
			if err := recover(); err != nil {
			}
		}()
		defer func() {
			if r.node.isLeader() {
				r.id++
			}
		}()
		r.mu.Lock()
		defer r.mu.Unlock()
		if _, ok := r.m[r.id]; ok {
			if len(r.m[r.id]) > 0 {
				go func(n *node, id uint64) {
					noOperationCommand := NewNoOperationCommand()
					if ok, _ := n.do(noOperationCommand, defaultCommandTimeout); ok != nil {
						r.reply(id, true)
						return
					}
					r.reply(id, false)
				}(r.node, r.id)
				return true
			}
		}
	}
	return false
}

func (r *readIndex) run() {
	for ch := range r.readChan {
		func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			if _, ok := r.m[r.id]; !ok {
				r.m[r.id] = []chan bool{}
			}
			r.m[r.id] = append(r.m[r.id], ch)
		}()
	}
}

func (r *readIndex) Stop() {
	if r.readChan != nil {
		close(r.readChan)
		r.readChan = nil
	}
}
