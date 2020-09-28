// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

func NewAddPeerCommand(nodeInfo *NodeInfo) Command {
	return &AddPeerCommand{
		NodeInfo: nodeInfo,
	}
}
func (c *AddPeerCommand) Type() int32 {
	return CommandTypeAddPeer
}

func (c *AddPeerCommand) Do(context interface{}) (interface{}, error) {
	node := context.(*Node)
	node.stateMachine.configuration.AddPeer(c.NodeInfo)
	node.stateMachine.configuration.load()
	return nil, nil
}
