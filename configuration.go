// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

type configuration struct {
	node    *node
	storage *ConfigurationStorage
	nodes   map[string]*NodeInfo
}

func newConfiguration(n *node) *configuration {
	c := &configuration{
		node:    n,
		storage: &ConfigurationStorage{},
		nodes:   make(map[string]*NodeInfo),
	}
	c.load()
	if _, ok := c.nodes[c.node.address]; !ok {
		c.AddPeer(&NodeInfo{Address: c.node.address})
	}
	return c
}

func (c *configuration) SetNodes(nodes []*NodeInfo) {
	for _, v := range nodes {
		c.nodes[v.Address] = v
	}
	c.save()
}
func (c *configuration) AddPeer(peer *NodeInfo) {
	c.nodes[peer.Address] = peer
	c.save()
}
func (c *configuration) RemovePeer(addr string) {
	if _, ok := c.nodes[addr]; ok {
		delete(c.nodes, addr)
	}
	c.save()
}
func (c *configuration) LookupPeer(addr string) *NodeInfo {
	if v, ok := c.nodes[addr]; ok {
		return v
	}
	return nil
}
func (c *configuration) Peers() []string {
	peers := make([]string, 0, len(c.nodes))
	for _, v := range c.nodes {
		if v.Address != c.node.address {
			peers = append(peers, v.Address)
		}
	}
	return peers
}
func (c *configuration) membership() []string {
	ms := make([]string, 0, len(c.nodes))
	for _, v := range c.nodes {
		if v.Address != c.node.address {
			ms = append(ms, fmt.Sprintf("%s;%t", v.Address, v.NonVoting))
		}
	}
	return ms
}
func (c *configuration) Nodes() []string {
	nodes := make([]string, 0)
	for _, v := range c.nodes {
		nodes = append(nodes, v.Address)
	}
	return nodes
}
func (c *configuration) save() {
	c.storage.Nodes = []*NodeInfo{}
	for _, v := range c.nodes {
		c.storage.Nodes = append(c.storage.Nodes, v)
	}
	b, _ := json.Marshal(c.storage)
	c.node.storage.OverWrite(defaultConfig, b)
}

func (c *configuration) load() error {
	if !c.node.storage.Exists(defaultConfig) {
		return nil
	}
	b, err := c.node.storage.Load(defaultConfig)
	if err != nil {
		return nil
	}
	if err = json.Unmarshal(b, c.storage); err != nil {
		return err
	}
	for _, v := range c.storage.Nodes {
		c.nodes[v.Address] = v
	}
	if !c.membershipChanges() {
		logger.Infof("configuration.reconfiguration !membershipChanges")
		return nil
	}
	lastNodesCount := c.node.NodesCount()
	c.node.deleteNotPeers(c.Peers())
	for _, v := range c.storage.Nodes {
		c.node.addNode(v)
	}
	nodesCount := c.node.NodesCount()
	if lastNodesCount != nodesCount {
		logger.Infof("configuration.load %s NodesCount %d==>%d", c.node.address, lastNodesCount, nodesCount)
	}
	return nil
}
func (c *configuration) reconfiguration() error {
	lastVotingsCount := c.node.votingsCount()
	c.node.consideredForMajorities()
	votingsCount := c.node.votingsCount()
	if lastVotingsCount != votingsCount {
		logger.Tracef("configuration.load %s VotingsCount %d==>%d", c.node.address, lastVotingsCount, votingsCount)
	}
	return nil
}

func (c *configuration) membershipChanges() bool {
	oldMembership := c.node.membership()
	newMembership := c.membership()
	sort.Strings(oldMembership)
	sort.Strings(newMembership)
	return !reflect.DeepEqual(oldMembership, newMembership)
}
