package main

import (
	"sync"
)
type Context struct {
	mut sync.RWMutex
	data  map[string]string
}

func NewContext() *Context {
	return &Context{data:make(map[string]string)}
}
func (ctx *Context) Set(key string, value string) {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()
	ctx.data[key] = value
}
func (ctx *Context) Get(key string) string {
	ctx.mut.RLock()
	defer ctx.mut.RUnlock()
	return ctx.data[key]
}

