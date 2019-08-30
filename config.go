package raft

import (
	"time"
)


const (
	DefaultNodeTick = time.Millisecond

	DefaultHearbeatTick = 100 * time.Millisecond
	DefaultDetectTick = 1000 * time.Millisecond

	DefaultRangeFactor = 5
	DefaultElectionTimeout = 1000 * time.Millisecond
	DefaultNormalOperationTimeout = 1000 * time.Millisecond
	DefaultHearbeatTimeout = 80 * time.Millisecond
	DefaultRequestVoteTimeout = 80 * time.Millisecond
	DefaultAppendEntriesTimeout = 1000 * time.Millisecond
	DefaultInstallSnapshotTimeout = 10000 * time.Millisecond
)

type Configuration struct {
	//CommitIndex uint64 	`json:"CommitIndex"`
	Name	string			`json:"Name"`
	Others []string			`json:"Others"`
}




