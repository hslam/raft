// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

func NewNoOperationCommand() Command {
	return &NoOperationCommand{}
}

func (c *NoOperationCommand) Type() int32 {
	return CommandTypeNoOperation
}

func (c *NoOperationCommand) UniqueID() string {
	return ""
}

func (c *NoOperationCommand) Do(context interface{}) (interface{}, error) {
	return true, nil
}
