package raft

import (
	"reflect"
	"sync"
)

type Command interface {
	Type() int32
	UniqueID() string
	Do(context interface{}) (interface{}, error)
}

type CommandType struct {
	mu    sync.RWMutex
	types map[int32]*sync.Pool
}

func (m *CommandType) register(cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.types[cmd.Type()]; ok {
		return ErrCommandTypeExisted
	}
	pool := &sync.Pool{New: func() interface{} {
		return reflect.New(reflect.Indirect(reflect.ValueOf(cmd)).Type()).Interface()
	}}
	pool.Put(pool.Get())
	m.types[cmd.Type()] = pool
	return nil
}

func (m *CommandType) clone(Type int32) Command {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if commandPool, ok := m.types[Type]; ok {
		return commandPool.Get().(Command)
	}
	Debugf("CommandType.clone Unregistered %d", Type)
	return nil
}

func (m *CommandType) put(cmd Command) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if commandPool, ok := m.types[cmd.Type()]; ok {
		commandPool.Put(cmd)
	}
}

func (m *CommandType) exists(cmd Command) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.types[cmd.Type()]
	return ok
}
