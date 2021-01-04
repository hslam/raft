// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"reflect"
	"sync"
)

const (
	noOperation              = 0x1
	addPeerOperation         = 0x2
	removePeerOperation      = 0x3
	reconfigurationOperation = 0x4
)

// Command represents a command.
type Command interface {
	// Type returns the command type. The type must be > 0.
	Type() uint64
	// Do executes the command with the context.
	Do(context interface{}) (reply interface{}, err error)
}

type commands struct {
	mu    sync.RWMutex
	types map[uint64]*sync.Pool
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

func (c *commands) clone(Type uint64) Command {
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

// Type implements the Command Type method.
func (c *DefaultCommand) Type() uint64 {
	return 0
}

// Do implements the Command Do method.
func (c *DefaultCommand) Do(context interface{}) (interface{}, error) {
	switch c.Operation {
	case noOperation:
		return true, nil
	case addPeerOperation:
		n := context.(*node)
		n.stateMachine.configuration.AddPeer(c.NodeInfo)
		n.stateMachine.configuration.load()
		return nil, nil
	case removePeerOperation:
		n := context.(*node)
		n.stateMachine.configuration.RemovePeer(c.NodeInfo.Address)
		return nil, nil
	case reconfigurationOperation:
		n := context.(*node)
		n.stateMachine.configuration.reconfiguration()
		return nil, nil
	}
	return nil, nil
}

var noOperationCommand = newNoOperationCommand()

// NewNoOperationCommand returns a new NoOperationCommand.
func newNoOperationCommand() Command {
	return &DefaultCommand{
		Operation: noOperation,
	}
}

var reconfigurationCommand = newReconfigurationCommand()

// NewReconfigurationCommand returns a new ReconfigurationCommand.
func newReconfigurationCommand() Command {
	return &DefaultCommand{
		Operation: reconfigurationOperation,
	}
}

// NewAddPeerCommand returns a new AddPeerCommand.
func newAddPeerCommand(nodeInfo *NodeInfo) Command {
	return &DefaultCommand{
		Operation: addPeerOperation,
		NodeInfo:  nodeInfo,
	}
}

// NewRemovePeerCommand returns a new RemovePeerCommand.
func newRemovePeerCommand(address string) Command {
	return &DefaultCommand{
		Operation: removePeerOperation,
		NodeInfo:  &NodeInfo{Address: address},
	}
}
