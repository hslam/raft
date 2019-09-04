package raft


type addPeerCommand struct {
	address string
}

func NewAddPeerCommand(address string) *addPeerCommand {
	return &addPeerCommand{
		address: address,
	}
}
func (c *addPeerCommand) Private()bool{
	return true
}
func (c *addPeerCommand) Type() uint32 {
	return CommandTypeAddPeer
}

func (c *addPeerCommand) Id() []byte {
	return append(uint32ToBytes(c.Type()),[]byte(c.address)...)
}

func (c *addPeerCommand) Do(node *Node)error{
	return nil
}
