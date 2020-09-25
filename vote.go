// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"strconv"
	"sync"
	"time"
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
	node      *Node
	vote      chan *vote
	voteDic   map[string]int
	voteTotal int
	voteCount int
	quorum    int
	total     int
	notice    chan bool
	timeout   time.Duration
}

func newVotes(node *Node) *votes {
	vs := &votes{
		node:   node,
		notice: make(chan bool, 1),
	}
	vs.Reset(1)
	vs.Clear()
	return vs
}

func (vs *votes) AddVote(v *vote) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if _, ok := vs.voteDic[v.Key()]; ok {
		return
	}
	vs.voteDic[v.Key()] = v.vote
	if v.term == vs.node.currentTerm.Id() {
		vs.voteTotal++
		vs.voteCount += v.vote
	}
	if vs.quorum > 0 && vs.voteCount >= vs.quorum {
		vs.notice <- true
	} else if vs.total > 0 && vs.voteTotal >= vs.total {
		vs.notice <- false
	}
}

func (vs *votes) Clear() {
	for {
		if len(vs.vote) > 0 {
			<-vs.vote
		} else {
			break
		}
	}
	vs.voteDic = make(map[string]int)
	vs.voteCount = 0
	vs.voteTotal = 0
}
func (vs *votes) Reset(count int) {
	if vs.vote != nil {
		close(vs.vote)
	}
	if count == 0 {
		count = 1
	}
	vs.vote = make(chan *vote, count)
}

func (vs *votes) Count() int {
	return vs.voteCount
}
func (vs *votes) Total() int {
	return vs.voteTotal
}
func (vs *votes) SetQuorum(quorum int) {
	vs.quorum = quorum
}
func (vs *votes) SetTotal(total int) {
	vs.total = total
}
func (vs *votes) GetNotice() chan bool {
	return vs.notice
}
func (vs *votes) SetTimeout(timeout time.Duration) {
	vs.timeout = timeout
	go vs.run()
}
func (vs *votes) run() {
	select {
	case <-time.After(vs.timeout):
		vs.notice <- false
	}
}
