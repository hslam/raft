package raft

func NewAddPeerCommand(nodeInfo *NodeInfo) Command {
	return &AddPeerCommand{
		NodeInfo:nodeInfo,
	}
}
func (c *AddPeerCommand) Type() int32 {
	return CommandTypeAddPeer
}

func (c *AddPeerCommand) UniqueID() string{
	return c.NodeInfo.Address
}

func (c *AddPeerCommand) Do(context interface{})(interface{},error){
	node := context.(*Node)
	node.stateMachine.configuration.AddPeer(c.NodeInfo)
	node.stateMachine.configuration.load()
	return nil, nil
}
