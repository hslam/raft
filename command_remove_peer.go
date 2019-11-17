package raft

type RemovePeerCommand struct {
	addr string
}

func NewRemovePeerCommand(address string) Command {
	return &RemovePeerCommand{
		addr:address,
	}
}
func (c *RemovePeerCommand) Type() int32 {
	return CommandTypeRemovePeer
}

func (c *RemovePeerCommand) UniqueID() string{
	return c.addr
}

func (c *RemovePeerCommand) Do(context interface{})(interface{},error){
	node := context.(*Node)
	node.stateMachine.configuration.RemovePeer(c.addr)
	return nil, nil
}
