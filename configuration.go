package raft

import (
	"encoding/json"
	"sort"
	"reflect"
)

type Configuration struct {
	node			*Node
	peers			[]string
	old				[]string
	isOldNew		bool
}
func newConfiguration(node *Node) *Configuration {
	c:=&Configuration{
		node:node,
	}
	c.load()
	return c
}

func (c *Configuration) SetPeers(peers []string){
	if len(c.peers)==0&&len(c.old)==0{
		c.peers=peers
		c.isOldNew=false
	}else if len(c.peers)==0&&len(c.old)>0{
		c.old=c.peers
		c.peers=peers
		c.isOldNew=true
	}
}
func (c *Configuration) ChangeToNew(){
	c.old=[]string{}
	c.isOldNew=false
}
func (c *Configuration) IsOldNew()bool {
	return c.isOldNew
}

func (c *Configuration) save() {
	peers := make([]string, len(c.node.peers))
	i := 0
	for _, client := range c.node.peers {
		peers[i] = client.address
		i++
	}
	c.peers=peers
	configurationStorage := &ConfigurationStorage{
		Address:c.node.address,
		Peers:c.peers,
	}
	b, _ := json.Marshal(configurationStorage)
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
	configurationStorage := &ConfigurationStorage{}
	if err = json.Unmarshal(b, configurationStorage); err != nil {
		return err
	}
	c.node.address=configurationStorage.Address
	peers:=configurationStorage.Peers
	if !c.isPeersChanged(peers){
		Tracef("is not PeersChanged %d %d",len(peers),len(c.peers))
		return nil
	}
	c.node.peers=make(map[string]*Peer)
	for _, client := range peers {
		c.node.addNode(client)
	}
	c.SetPeers(peers)
	return nil
}

func  (c *Configuration) isPeersChanged(peers []string)  bool {
	old_nodes:=append(c.peers[:],c.node.address)
	new_nodes:=append(peers[:],c.node.address)
	sort.Strings(old_nodes)
	sort.Strings(new_nodes)
	return !reflect.DeepEqual(old_nodes, new_nodes)
}