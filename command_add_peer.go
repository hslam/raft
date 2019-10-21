package raft


type addPeerCommand struct {
	address string
}

func NewAddPeerCommand(address string) Command {
	return &addPeerCommand{
		address: address,
	}
}
func (c *addPeerCommand) Type() int32 {
	return CommandTypeAddPeer
}

func (c *addPeerCommand) UniqueID() string{
	return c.address
}

func (c *addPeerCommand) Do(context interface{})(interface{},error){
	return nil, nil
}
