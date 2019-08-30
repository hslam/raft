package raft

import (
	"time"
)


const (
	DefaultHearbeatTick = 100 * time.Millisecond
	DefaultDetectTick = 1000 * time.Millisecond

	DefaultRangeFactor = 5

	DefaultWaitHearbeatTimeout = 1000 * time.Millisecond
	DefaultElectionTimeout = 1000 * time.Millisecond
	DefaultHearbeatTimeout = 100 * time.Millisecond
	DefaultRequestVoteTimeout = 150 * time.Millisecond
	DefaultAppendEntriesTimeout = 1000 * time.Millisecond
	DefaultInstallSnapshotTimeout = 10000 * time.Millisecond
)

type Configuration struct {
	//CommitIndex uint64 	`json:"CommitIndex"`
	Node string			`json:"Node"`
	Peers []string		`json:"Peers"`
}




