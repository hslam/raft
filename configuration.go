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
		Name:c.node.address,
		Peers:c.peers,
	}
	b, _ := json.Marshal(configurationStorage)
	c.node.storage.SafeOverWrite(DefaultConfig,b)
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
	c.node.address=configurationStorage.Name
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
	//n.log.updateCommitIndex(config.CommitIndex)
	return nil
}

func  (c *Configuration) isPeersChanged(peers []string)  bool {
	if len(peers) == 0 && len(c.peers) == 0 {
		return false
	}
	sort.Strings(peers)
	sort.Strings(c.peers)
	return !reflect.DeepEqual(peers, c.peers)
}