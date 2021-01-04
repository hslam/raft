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
	Type() int64
	// Do does the command.
	Do(context interface{}) (interface{}, error)
}

type commands struct {
	mu    sync.RWMutex
	types map[int64]*sync.Pool
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

func (c *commands) clone(Type int64) Command {
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

// NewAddPeerCommand returns a new AddPeerCommand.
func newAddPeerCommand(nodeInfo *NodeInfo) Command {
	return &AddPeerCommand{
		NodeInfo: nodeInfo,
	}
}

// Type implements the Command Type method.
func (c *AddPeerCommand) Type() int64 {
	return commandTypeAddPeer
}

// Do implements the Command Do method.
func (c *AddPeerCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.AddPeer(c.NodeInfo)
	n.stateMachine.configuration.load()
	return nil, nil
}

var noOperationCommand = newNoOperationCommand()

// NewNoOperationCommand returns a new NoOperationCommand.
func newNoOperationCommand() Command {
	return &NoOperationCommand{}
}

// Type implements the Command Type method.
func (c *NoOperationCommand) Type() int64 {
	return commandTypeNoOperation
}

// Do implements the Command Do method.
func (c *NoOperationCommand) Do(context interface{}) (interface{}, error) {
	return true, nil
}

// NewReconfigurationCommand returns a new ReconfigurationCommand.
func newReconfigurationCommand() Command {
	return &ReconfigurationCommand{}
}

// Type implements the Command Type method.
func (c *ReconfigurationCommand) Type() int64 {
	return commandTypeReconfiguration
}

// Do implements the Command Do method.
func (c *ReconfigurationCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.reconfiguration()
	return nil, nil
}

// NewRemovePeerCommand returns a new RemovePeerCommand.
func newRemovePeerCommand(address string) Command {
	return &RemovePeerCommand{
		Address: address,
	}
}

// Type implements the Command Type method.
func (c *RemovePeerCommand) Type() int64 {
	return commandTypeRemovePeer
}

// Do implements the Command Do method.
func (c *RemovePeerCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.RemovePeer(c.Address)
	return nil, nil
}
