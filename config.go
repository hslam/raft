package raft

import (
	"time"
)

const (
	DefaultStartWait = 3*1000*time.Millisecond
	DefaultNodeTick = time.Millisecond
	DefaultNodeTracePrintTick = 1000 * time.Millisecond
	DefaultCompactionTick 	=	60*1000 * time.Millisecond
	DefaultHeartbeatTick = 100 * time.Millisecond
	DefaultDetectTick = 100 * time.Millisecond
	DefaultKeepAliveTick = 10000 * time.Millisecond
	DefaultElectionTimeout = 1000 * time.Millisecond
	DefaultHearbeatTimeout = 100 * time.Millisecond
	DefaultRequestVoteTimeout = 1000 * time.Millisecond
	DefaultAppendEntriesTimeout = 10*1000 * time.Millisecond
	DefaultInstallSnapshotTimeout = 10*1000 * time.Millisecond

	DefaultCommandTimeout = 60*1000 * time.Millisecond

	DefaultMaxConcurrency = 1024*32
	DefaultMaxBatch = 1024*32
	DefaultMaxCacheEntries=1024*32
	DefaultMaxDelay	=	  time.Millisecond
	DefaultCheckDelay	=	time.Millisecond*5
	DefaultCommitDelay	=	 time.Millisecond*5
	DefaultRetryTimes	=	5
	DefaultLeaseFactor = 0.9

	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultLog = "log"
	DefaultCompaction = "compaction"
	DefaultIndex = "index"
	DefaultLastLogIndex = "lastlogindex"
	DefaultCommitIndex = "commitindex"
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



