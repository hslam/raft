// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"strconv"
	"sync"
)

type vote struct {
	id   string
	term uint64
	vote int
}

func newVote(id string, term uint64, v int) *vote {
	return &vote{
		id:   id,
		term: term,
		vote: v,
	}
}

func (v *vote) Key() string {
	return v.id + strconv.FormatUint(v.term, 10)
}

type votes struct {
	mu        sync.RWMutex
	node      *node
	vote      chan *vote
	voteDic   map[string]int
	voteCount int
}

func newVotes(n *node) *votes {
	vs := &votes{
		node: n,
	}
	vs.Reset(1)
	vs.Clear()
	return vs
}

func (vs *votes) AddVote(v *vote) {
	vs.mu.Lock()
	if _, ok := vs.voteDic[v.Key()]; ok {
		vs.mu.Unlock()
		return
	}
	vs.voteDic[v.Key()] = v.vote
	if v.term == vs.node.currentTerm.Load() {
		vs.voteCount += v.vote
	}
	vs.mu.Unlock()
}

func (vs *votes) Clear() {
	vs.mu.Lock()
	clearVote(vs.vote)
	vs.voteDic = make(map[string]int)
	vs.voteCount = 0
	vs.mu.Unlock()

}

func (vs *votes) Reset(count int) {
	vs.mu.Lock()
	if vs.vote != nil {
		close(vs.vote)
	}
	if count == 0 {
		count = 1
	}
	vs.vote = make(chan *vote, count)
	vs.mu.Unlock()
}

func (vs *votes) Count() int {
	vs.mu.Lock()
	voteCount := vs.voteCount
	vs.mu.Unlock()
	return voteCount
}

func clearVote(v chan *vote) {
	for len(v) > 0 {
		clearVoteOnce(v)
	}
}

func clearVoteOnce(v chan *vote) {
	select {
	case <-v:
	default:
	}
}
