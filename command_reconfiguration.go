// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

func NewReconfigurationCommand() Command {
	return &ReconfigurationCommand{}
}
func (c *ReconfigurationCommand) Type() int32 {
	return CommandTypeReconfiguration
}

func (c *ReconfigurationCommand) UniqueID() string {
	return ""
}

func (c *ReconfigurationCommand) Do(context interface{}) (interface{}, error) {
	node := context.(*Node)
	node.stateMachine.configuration.reconfiguration()
	return nil, nil
}
