package raft

type AddPeerCommand struct {
	nodeInfo *NodeInfo
}

func NewAddPeerCommand(address string,data []byte) Command {
	return &AddPeerCommand{
		nodeInfo:&NodeInfo{Address:address,Data:data},
	}
}
func (c *AddPeerCommand) Type() int32 {
	return CommandTypeAddPeer
}

func (c *AddPeerCommand) UniqueID() string{
	return c.nodeInfo.Address
}

func (c *AddPeerCommand) Do(context interface{})(interface{},error){
	node := context.(*Node)
	node.stateMachine.configuration.AddPeer(c.nodeInfo)
	return nil, nil
}
