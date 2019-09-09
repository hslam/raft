package raft

import (
	"sync"
	"errors"
)

const VoteForPrefix  = "VoteFor "

type VoteFor struct {
	mu 								sync.RWMutex
	node 							*Node
	address							string
}
func newVoteFor(node *Node) *VoteFor {
	voteFor:=&VoteFor{
		node:node,
	}
	voteFor.loadVoteFor()
	return voteFor
}
func (voteFor *VoteFor) Reset() {
	voteFor.mu.Lock()
	defer voteFor.mu.Unlock()
	voteFor.address=""
	voteFor.saveVoteFor()
}
func (voteFor *VoteFor) Set(address string) {
	voteFor.mu.Lock()
	defer voteFor.mu.Unlock()
	voteFor.address=address
	voteFor.saveVoteFor()
}

func (voteFor *VoteFor) String()string {
	voteFor.mu.RLock()
	defer voteFor.mu.RUnlock()
	return voteFor.address
}

func (voteFor *VoteFor) saveVoteFor() {
	voteFor.node.storage.SafeOverWrite(DefaultVoteFor,[]byte(VoteForPrefix+voteFor.address))
}

func (voteFor *VoteFor) loadVoteFor() error {
	if !voteFor.node.storage.Exists(DefaultVoteFor){
		return errors.New("votefor file is not existed")
	}
	b, err := voteFor.node.storage.Load(DefaultVoteFor)
	if err != nil {
		return err
	}
	data:=b[len(VoteForPrefix):]
	voteFor.address = string(data)
	return nil
}