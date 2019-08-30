package raft

import "strconv"

type Vote struct {
	candidateId	string
	term		uint64
	vote		int
}

func (v *Vote)Key() string {
	return v.candidateId+strconv.FormatUint(v.term,10)
}