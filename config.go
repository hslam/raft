package raft

import (
	"time"
)


const (
	DefaultNodeTick = time.Millisecond
	DefaultCompactionTick 	=	60*1000 * time.Millisecond
	DefaultHearbeatTick = 50 * time.Millisecond
	DefaultDetectTick = 100 * time.Millisecond

	DefaultElectionTimeout = 150 * time.Millisecond

	DefaultHearbeatTimeout = 50 * time.Millisecond
	DefaultRequestVoteTimeout = 100 * time.Millisecond
	DefaultAppendEntriesTimeout = 1000 * time.Millisecond
	DefaultInstallSnapshotTimeout = 10000 * time.Millisecond

	DefaultMaxKeyLength = 64
	DefaultCommandTimeout = 10000 * time.Millisecond

	DefaultMaxConcurrency = 256
	DefaultMaxBatch = 256
	DefaultMaxCacheEntries=1024
	DefaultMaxDelay	=	100 * time.Millisecond
	CommandTypeAddPeer =1
	CommandTypeRemovePeer =2

)




