// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

// NewRemovePeerCommand returns a new RemovePeerCommand.
func NewRemovePeerCommand(address string) Command {
	return &RemovePeerCommand{
		Address: address,
	}
}

// Type implements the Command Type method.
func (c *RemovePeerCommand) Type() int32 {
	return commandTypeRemovePeer
}

// Do implements the Command Do method.
func (c *RemovePeerCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.RemovePeer(c.Address)
	return nil, nil
}
