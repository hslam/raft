// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

// NewNoOperationCommand returns a new NoOperationCommand.
func NewNoOperationCommand() Command {
	return &NoOperationCommand{}
}

// Type implements the Command Type method.
func (c *NoOperationCommand) Type() int32 {
	return commandTypeNoOperation
}

// Do implements the Command Do method.
func (c *NoOperationCommand) Do(context interface{}) (interface{}, error) {
	return true, nil
}
