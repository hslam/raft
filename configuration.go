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
	node    *node
	mu      sync.Mutex
	members map[string]*Member
}

func newConfiguration(n *node) *configuration {
	c := &configuration{
		node:    n,
		members: make(map[string]*Member),
	}
	c.load()
	if _, ok := c.members[c.node.address]; !ok {
		c.AddMember(&Member{Address: c.node.address})
	}
	return c
}

func (c *configuration) SetMembers(members []*Member) {
	c.mu.Lock()
	for _, v := range members {
		c.members[v.Address] = v
	}
	c.mu.Unlock()
	c.save()
}

func (c *configuration) AddMember(member *Member) {
	c.mu.Lock()
	c.members[member.Address] = member
	c.mu.Unlock()
	c.save()
	if c.node.memberChange != nil {
		go c.node.memberChange()
	}
}

func (c *configuration) RemoveMember(addr string) {
	c.mu.Lock()
	if _, ok := c.members[addr]; ok {
		delete(c.members, addr)
	}
	c.mu.Unlock()
	c.save()
	if c.node.memberChange != nil {
		go c.node.memberChange()
	}
}

func (c *configuration) LookupMember(addr string) (member *Member) {
	c.mu.Lock()
	if v, ok := c.members[addr]; ok {
		member = v
	}
	c.mu.Unlock()
	return
}

func (c *configuration) Peers() []string {
	c.mu.Lock()
	peers := make([]string, 0, len(c.members))
	for _, v := range c.members {
		if v.Address != c.node.address {
			peers = append(peers, v.Address)
		}
	}
	c.mu.Unlock()
	return peers
}

func (c *configuration) Members() []*Member {
	c.mu.Lock()
	members := make([]*Member, 0, len(c.members))
	for _, v := range c.members {
		member := &Member{}
		*member = *v
		members = append(members, member)
	}
	c.mu.Unlock()
	return members
}

func (c *configuration) membership() []string {
	c.mu.Lock()
	ms := make([]string, 0, len(c.members))
	for _, v := range c.members {
		ms = append(ms, fmt.Sprintf("%s;%t", v.Address, v.NonVoting))
	}
	c.mu.Unlock()
	return ms
}

func (c *configuration) save() {
	n := nodes{}
	c.mu.Lock()
	for _, v := range c.members {
		n = append(n, v)
	}
	sort.Sort(n)
	storage := &ConfigurationStorage{Members: n}
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
	for _, v := range storage.Members {
		c.members[v.Address] = v
	}
	c.mu.Unlock()
	if !c.membershipChanges() {
		//logger.Tracef("configuration.load !membershipChanges")
		return nil
	}
	lastNodesCount := c.node.NodesCount()
	c.node.deleteNotPeers(c.Peers())
	for _, v := range storage.Members {
		c.node.addNode(v)
	}
	nodesCount := c.node.NodesCount()
	if lastNodesCount != nodesCount {
		c.node.logger.Tracef("configuration.load %s NodesCount %d==>%d", c.node.address, lastNodesCount, nodesCount)
	}
	return nil
}

func (c *configuration) reset() {
	c.mu.Lock()
	c.node.storage.Truncate(defaultConfig, 0)
	c.members = make(map[string]*Member)
	c.mu.Unlock()
}

func (c *configuration) reconfiguration() error {
	lastVotingsCount := c.node.votingsCount()
	c.node.consideredForMajorities()
	votingsCount := c.node.votingsCount()
	if lastVotingsCount != votingsCount {
		c.node.logger.Tracef("configuration.reconfiguration %s VotingsCount %d==>%d", c.node.address, lastVotingsCount, votingsCount)
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

type nodes []*Member

func (n nodes) Len() int           { return len(n) }
func (n nodes) Less(i, j int) bool { return n[i].Address < n[j].Address }
func (n nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
