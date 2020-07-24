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
			return &Invoker{}
		},
	}
	invokerPool.Put(invokerPool.Get())
	donePool = &sync.Pool{
		New: func() interface{} {
			return make(chan *Invoker, 10)
		},
	}
	donePool.Put(donePool.Get())
}

type Invoker struct {
	index   uint64
	Command Command
	Reply   interface{}
	Error   error
	Done    chan *Invoker
}

func (invoker *Invoker) done() {
	select {
	case invoker.Done <- invoker:
	default:
	}
}
