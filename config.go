package raft

import (
	"time"
)


const (
	DefaultStartWait = 10*1000*time.Microsecond
	DefaultNodeTick = time.Microsecond*100
	DefaultCompactionTick 	=	60*1000 * time.Millisecond
	DefaultHearbeatTick = 100 * time.Millisecond
	DefaultDetectTick = 100 * time.Millisecond
	DefaultKeepAliveTick = 10000 * time.Millisecond
	DefaultElectionTimeout = 1000 * time.Millisecond
	DefaultHearbeatTimeout = 100 * time.Millisecond
	DefaultRequestVoteTimeout = 1000 * time.Millisecond
	DefaultAppendEntriesTimeout = 1000 * time.Millisecond
	DefaultInstallSnapshotTimeout = 10*1000 * time.Millisecond

	DefaultCommandTimeout = 10*1000 * time.Millisecond

	DefaultMaxConcurrency = 1024*64
	DefaultMaxBatch = 1024*64
	DefaultMaxCacheEntries=1024*1024
	DefaultMaxDelay	=	 time.Millisecond

	DefaultRetryTimes	=	5

	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultLog = "log"
	DefaultCompaction = "compaction"
	DefaultIndex = "index"
	DefaultTerm = "term"
	DefaultVoteFor = "votefor"
	DefaultSnapshot = "snapshot"
	DefaultMd5 = "md5"
	DefaultLastSaved = "lastsaved"
	DefaultTmp = ".tmp"

	CommandTypeAddPeer =-1
	CommandTypeRemovePeer =-2

)
type SnapshotSyncType int

const (
	Never SnapshotSyncType = 0
	EverySecond SnapshotSyncType= 1
	Always SnapshotSyncType= 2
	DefaultMaxTimesSaveSnapshot = 256
)



