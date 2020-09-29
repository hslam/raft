// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"reflect"
	"sync"
)

// Command represents a command.
type Command interface {
	// Type returns the command type. The type must be >= 0.
	Type() int32
	// Do does the command.
	Do(context interface{}) (interface{}, error)
}

type commands struct {
	mu    sync.RWMutex
	types map[int32]*sync.Pool
}

func (c *commands) register(cmd Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.types[cmd.Type()]; ok {
		return ErrCommandTypeExisted
	}
	pool := &sync.Pool{New: func() interface{} {
		return reflect.New(reflect.Indirect(reflect.ValueOf(cmd)).Type()).Interface()
	}}
	pool.Put(pool.Get())
	c.types[cmd.Type()] = pool
	return nil
}

func (c *commands) clone(Type int32) Command {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if commandPool, ok := c.types[Type]; ok {
		return commandPool.Get().(Command)
	}
	logger.Debugf("CommandType.clone Unregistered %d", Type)
	return nil
}

func (c *commands) put(cmd Command) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if commandPool, ok := c.types[cmd.Type()]; ok {
		commandPool.Put(cmd)
	}
}

func (c *commands) exists(cmd Command) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.types[cmd.Type()]
	return ok
}
