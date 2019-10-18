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
	DefaultCheckLogTick	=	 time.Second
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
	DefaultTarTick	=	 time.Minute

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
	DefaultLastIncludedIndex = "lastincludedindex"
	DefaultLastIncludedTerm = "lastincludedterm"
	DefaultLastTarIndex = "lasttarindex"


	DefaultMd5 = "md5"
	DefaultTar = "tar"
	DefaultTarGz = "tar.gz"
	DefaultTmp = ".tmp"
	DefaultReadFileBufferSize= 1 << 24
	CommandTypeAddPeer =-1
	CommandTypeRemovePeer =-2

	DefaultNumInstallSnapshot = 1<<27
	DefaultMaxEntriesPerFile = 1 << 27
	DefaultChunkSize int64 = 1 << 24

)
type SnapshotSyncType int

const (
	Never SnapshotSyncType = 0
	EverySecond SnapshotSyncType= 1
	EveryMinute SnapshotSyncType= 2
	EveryHour SnapshotSyncType= 3
	DefalutSave SnapshotSyncType= 4
	Save SnapshotSyncType= 5
	Always SnapshotSyncType= 9

	SecondsSaveSnapshot1 = 900
	ChangesSaveSnapshot1 = 1
	SecondsSaveSnapshot2 = 300
	ChangesSaveSnapshot2 = 10
	SecondsSaveSnapshot3 = 60
	ChangesSaveSnapshot3 = 10000
)



