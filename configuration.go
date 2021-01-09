// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type configuration struct {
	node  *node
	mu    sync.Mutex
	nodes map[string]*NodeInfo
}

func newConfiguration(n *node) *configuration {
	c := &configuration{
		node:  n,
		nodes: make(map[string]*NodeInfo),
	}
	c.load()
	if _, ok := c.nodes[c.node.address]; !ok {
		c.AddPeer(&NodeInfo{Address: c.node.address})
	}
	return c
}

func (c *configuration) SetNodes(nodes []*NodeInfo) {
	c.mu.Lock()
	for _, v := range nodes {
		c.nodes[v.Address] = v
	}
	c.mu.Unlock()
	c.save()
}

func (c *configuration) AddPeer(peer *NodeInfo) {
	c.mu.Lock()
	c.nodes[peer.Address] = peer
	c.mu.Unlock()
	c.save()
	if c.node.memberChange != nil {
		go c.node.memberChange()
	}
}

func (c *configuration) RemovePeer(addr string) {
	c.mu.Lock()
	if _, ok := c.nodes[addr]; ok {
		delete(c.nodes, addr)
	}
	c.mu.Unlock()
	c.save()
	if c.node.memberChange != nil {
		go c.node.memberChange()
	}
}

func (c *configuration) LookupPeer(addr string) *NodeInfo {
	c.mu.Lock()
	if v, ok := c.nodes[addr]; ok {
		c.mu.Unlock()
		return v
	}
	c.mu.Unlock()
	return nil
}

func (c *configuration) Peers() []string {
	c.mu.Lock()
	peers := make([]string, 0, len(c.nodes))
	for _, v := range c.nodes {
		if v.Address != c.node.address {
			peers = append(peers, v.Address)
		}
	}
	c.mu.Unlock()
	return peers
}

func (c *configuration) membership() []string {
	c.mu.Lock()
	ms := make([]string, 0, len(c.nodes))
	for _, v := range c.nodes {
		ms = append(ms, fmt.Sprintf("%s;%t", v.Address, v.NonVoting))
	}
	c.mu.Unlock()
	return ms
}

func (c *configuration) save() {
	n := nodes{}
	c.mu.Lock()
	for _, v := range c.nodes {
		n = append(n, v)
	}
	sort.Sort(n)
	storage := &ConfigurationStorage{Nodes: n}
	c.mu.Unlock()
	b, _ := json.Marshal(storage)
	c.node.storage.OverWrite(defaultConfig, b)
}

func (c *configuration) load() error {
	b, err := c.node.storage.Load(defaultConfig)
	if err != nil {
		return err
	}
	storage := &ConfigurationStorage{}
	json.Unmarshal(b, storage)
	c.mu.Lock()
	for _, v := range storage.Nodes {
		c.nodes[v.Address] = v
	}
	c.mu.Unlock()
	if !c.membershipChanges() {
		//logger.Tracef("configuration.load !membershipChanges")
		return nil
	}
	lastNodesCount := c.node.NodesCount()
	c.node.deleteNotPeers(c.Peers())
	for _, v := range storage.Nodes {
		c.node.addNode(v)
	}
	nodesCount := c.node.NodesCount()
	if lastNodesCount != nodesCount {
		logger.Tracef("configuration.load %s NodesCount %d==>%d", c.node.address, lastNodesCount, nodesCount)
	}
	return nil
}

func (c *configuration) reconfiguration() error {
	lastVotingsCount := c.node.votingsCount()
	c.node.consideredForMajorities()
	votingsCount := c.node.votingsCount()
	if lastVotingsCount != votingsCount {
		logger.Tracef("configuration.reconfiguration %s VotingsCount %d==>%d", c.node.address, lastVotingsCount, votingsCount)
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

type nodes []*NodeInfo

func (n nodes) Len() int           { return len(n) }
func (n nodes) Less(i, j int) bool { return n[i].Address < n[j].Address }
func (n nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
