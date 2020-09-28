// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

// NewAddPeerCommand returns a new AddPeerCommand.
func NewAddPeerCommand(nodeInfo *NodeInfo) Command {
	return &AddPeerCommand{
		NodeInfo: nodeInfo,
	}
}

// Type implements the Command Type method.
func (c *AddPeerCommand) Type() int32 {
	return commandTypeAddPeer
}

// Do implements the Command Do method.
func (c *AddPeerCommand) Do(context interface{}) (interface{}, error) {
	n := context.(*node)
	n.stateMachine.configuration.AddPeer(c.NodeInfo)
	n.stateMachine.configuration.load()
	return nil, nil
}
