// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

// NewReconfigurationCommand returns a new ReconfigurationCommand.
func newReconfigurationCommand() Command {
	return &ReconfigurationCommand{}
}

// Type implements the Command Type method.
func (c *ReconfigurationCommand) Type() int32 {
	return commandTypeReconfiguration
}

// Do implements the Command Do method.
func (c *ReconfigurationCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.reconfiguration()
	return nil, nil
}
