package raft


type Client struct {
	node					*Node
	address					string
	alive					bool
	follower				bool
}

func newClient(node *Node, address string) *Client {
	c:=&Client{
		node :					node,
		address :				address,
	}
	return c
}

func (c *Client) heartbeat() {
	c.alive,c.follower=c.node.raft.Heartbeat(c.address)
}

func (c *Client) requestVote() {
	c.alive=c.node.raft.RequestVote(c.address)
}
func (c *Client) appendEntries() {
	c.alive=c.node.raft.AppendEntries(c.address)
}
func (c *Client) installSnapshot() {
	c.alive=c.node.raft.InstallSnapshot(c.address)
}

func (c *Client) ping() {
	c.alive=c.node.rpcs.Ping(c.address)
	//Debugf("Node.ping %s -> %s %t",c.server.address,c.address,c.ping)
}
