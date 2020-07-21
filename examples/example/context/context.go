package context

import (
	"sync"
)

type Context struct {
	mut  sync.RWMutex
	data string
}

func NewContext() *Context {
	return &Context{data: ""}
}
func (ctx *Context) Set(value string) {
	ctx.mut.Lock()
	defer ctx.mut.Unlock()
	ctx.data = value
}
func (ctx *Context) Get() string {
	ctx.mut.RLock()
	defer ctx.mut.RUnlock()
	return ctx.data
}
