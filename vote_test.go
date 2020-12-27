package raft

import (
	"github.com/hslam/atomic"
	"testing"
)

func TestClearVote(t *testing.T) {
	vs := make(chan *vote, 1)
	vs <- &vote{}
	clearVote(vs)
	if len(vs) > 0 {
		t.Error()
	}
}

func TestAddVote(t *testing.T) {
	vs := newVotes(&node{currentTerm: atomic.NewUint64(1)})
	v := &vote{term: 1}
	vs.AddVote(v)
	vs.AddVote(v)
	if vs.voteCount > 1 {
		t.Error(vs.voteCount)
	}
}
