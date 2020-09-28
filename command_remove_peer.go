// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import "fmt"

func NewRemovePeerCommand(address string) Command {
	return &RemovePeerCommand{
		Address: address,
	}
}

func (c *RemovePeerCommand) Type() int32 {
	return CommandTypeRemovePeer
}

func (c *RemovePeerCommand) Do(context interface{}) (interface{}, error) {
	fmt.Println("ReconfigurationCommand")
	return nil, nil
	node := context.(*Node)
	node.stateMachine.configuration.RemovePeer(c.Address)
	return nil, nil
}
