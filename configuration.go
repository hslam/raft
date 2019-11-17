package raft

import (
	"encoding/json"
	"sort"
	"reflect"
)

type Configuration struct {
	node			*Node
	storage 		*ConfigurationStorage
	nodes 			map[string]*NodeInfo
}
func newConfiguration(node *Node) *Configuration {
	c:=&Configuration{
		node:node,
		storage:&ConfigurationStorage{},
		nodes:make(map[string]*NodeInfo),
	}
	c.load()
	return c
}

func (c *Configuration) SetNodes(nodes []*NodeInfo){
	c.storage.Nodes=append(c.storage.Nodes,nodes...)
	for _,v:=range nodes{
		c.nodes[v.Address]=v
	}
}
func (c *Configuration) AddPeer(peer *NodeInfo){
	c.storage.Nodes=append(c.storage.Nodes,peer)
	c.nodes[peer.Address]=peer
}
func (c *Configuration) RemovePeer(addr string){
	for i,v:=range c.storage.Nodes{
		if v.Address==addr{
			c.storage.Nodes = append(c.storage.Nodes[:i], c.storage.Nodes[i+1:]...)
		}
	}
	if _,ok:=c.nodes[addr];ok{
		delete(c.nodes,addr)
	}
}
func (c *Configuration) LookupPeer(addr string)*NodeInfo{
	if v,ok:=c.nodes[addr];ok{
		return v
	}
	return nil
}
func (c *Configuration) Peers()[] string{
	peers:=make([]string,0)
	for _,v:=range c.storage.Nodes{
		if v.Address!=c.node.address{
			peers=append(peers,v.Address)
		}
	}
	return peers
}
func (c *Configuration) save() {
	b, _ := json.Marshal(c.storage)
	c.node.storage.OverWrite(DefaultConfig,b)
}

func (c *Configuration) load() error {
	if !c.node.storage.Exists(DefaultConfig){
		return nil
	}
	b, err := c.node.storage.Load(DefaultConfig)
	if err != nil {
		return nil
	}
	if err = json.Unmarshal(b, c.storage); err != nil {
		return err
	}
	for _,v:=range c.storage.Nodes{
		c.nodes[v.Address]=v
	}
	if !c.membershipChanges(){
		Traceln("Configuration.load !membershipChanges")
		return nil
	}
	c.node.deleteNotPeers(c.Peers())
	for _, v := range c.storage.Nodes {
		c.node.addNode(v.Address)
	}
	return nil
}

func  (c *Configuration) membershipChanges()bool {
	old_peers:=c.node.Peers()
	new_peers:=c.Peers()
	sort.Strings(old_peers)
	sort.Strings(new_peers)
	return !reflect.DeepEqual(old_peers, new_peers)
}