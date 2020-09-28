// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
)

var (
	invokerPool *sync.Pool
	donePool    *sync.Pool
)

func init() {
	invokerPool = &sync.Pool{
		New: func() interface{} {
			return &invoker{}
		},
	}
	invokerPool.Put(invokerPool.Get())
	donePool = &sync.Pool{
		New: func() interface{} {
			return make(chan *invoker, 10)
		},
	}
	donePool.Put(donePool.Get())
}

type invoker struct {
	index   uint64
	Command Command
	Reply   interface{}
	Error   error
	Done    chan *invoker
}

func (i *invoker) done() {
	select {
	case i.Done <- i:
	default:
	}
}
