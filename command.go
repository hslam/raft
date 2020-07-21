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
	types map[int32]Command
}

func (m *CommandType) register(cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.types[cmd.Type()]; ok {
		return ErrCommandTypeExisted
	}
	m.types[cmd.Type()] = cmd
	return nil
}

func (m *CommandType) clone(Type int32) Command {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if command, ok := m.types[Type]; ok {
		command_copy := reflect.New(reflect.Indirect(reflect.ValueOf(command)).Type()).Interface()
		return command_copy.(Command)
	}
	Debugf("CommandType.clone Unregistered %d", Type)
	return nil
}

func (m *CommandType) exists(cmd Command) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.types[cmd.Type()]
	return ok
}
