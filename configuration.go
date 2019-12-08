package raft

import (
	"encoding/json"
	"sort"
	"reflect"
	"fmt"
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
	if _,ok:=c.nodes[c.node.address];!ok{
		c.AddPeer(&NodeInfo{Address:c.node.address})
	}
	return c
}

func (c *Configuration) SetNodes(nodes []*NodeInfo){
	for _,v:=range nodes{
		c.nodes[v.Address]=v
	}
	c.save()
}
func (c *Configuration) AddPeer(peer *NodeInfo){
	c.nodes[peer.Address]=peer
	c.save()
}
func (c *Configuration) RemovePeer(addr string){
	if _,ok:=c.nodes[addr];ok{
		delete(c.nodes,addr)
	}
	c.save()
}
func (c *Configuration) LookupPeer(addr string)*NodeInfo{
	if v,ok:=c.nodes[addr];ok{
		return v
	}
	return nil
}
func (c *Configuration) Peers()[] string{
	peers:=make([]string,0,len(c.nodes))
	for _,v:=range c.nodes{
		if v.Address!=c.node.address{
			peers=append(peers,v.Address)
		}
	}
	return peers
}
func (c *Configuration) membership()[] string{
	ms:=make([]string,0,len(c.nodes))
	for _,v:=range c.nodes{
		if v.Address!=c.node.address{
			ms=append(ms,fmt.Sprintf("%s;%t",v.Address,v.NonVoting))
		}
	}
	return ms
}
func (c *Configuration) Nodes()[] string{
	nodes:=make([]string,0)
	for _,v:=range c.nodes{
		nodes=append(nodes,v.Address)
	}
	return nodes
}
func (c *Configuration) save() {
	c.storage.Nodes=[]*NodeInfo{}
	for _,v:=range c.nodes{
		c.storage.Nodes=append(c.storage.Nodes,v)
	}
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
		Infof("Configuration.reconfiguration !membershipChanges")
		return nil
	}
	lastNodesCount:=c.node.NodesCount()
	c.node.deleteNotPeers(c.Peers())
	for _, v := range c.storage.Nodes {
		c.node.addNode(v)
	}
	nodesCount:=c.node.NodesCount()
	if lastNodesCount!=nodesCount{
		Infof("Configuration.load %s NodesCount %d==>%d",c.node.address,lastNodesCount,nodesCount)
	}
	return nil
}
func (c *Configuration)reconfiguration() error {
	lastVotingsCount:=c.node.votingsCount()
	c.node.consideredForMajorities()
	votingsCount:=c.node.votingsCount()
	if lastVotingsCount!=votingsCount{
		Tracef("Configuration.load %s VotingsCount %d==>%d",c.node.address,lastVotingsCount,votingsCount)
	}
	return nil
}

func  (c *Configuration) membershipChanges()bool {
	old_ms:=c.node.membership()
	new_ms:=c.membership()
	sort.Strings(old_ms)
	sort.Strings(new_ms)
	return !reflect.DeepEqual(old_ms, new_ms)
}

