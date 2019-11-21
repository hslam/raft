package raft

import (
	"time"
)

const (
	DefaultStartWait = 3*1000*time.Millisecond
	DefaultNodeTick = time.Millisecond*100
	DefaultUpdateTick = time.Millisecond*100
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
	DefaultInstallSnapshotTimeout = 60*1000 * time.Millisecond

	DefaultCommandTimeout = 60*1000 * time.Millisecond
	DefaultMaxConcurrencyRead = 1024*1024
	DefaultMaxConcurrency = 1024*1024
	DefaultMaxCacheEntries=1024*1024
	DefaultMaxBatch = 1024*1024
	DefaultTarTick	=	 time.Hour
	DefaultMaxDelay = time.Millisecond

	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultLog = "log"
	DefaultCompaction = "compaction"
	DefaultIndex = "index"
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
	DefaultFlush = ".flush"
	DefaultReadFileBufferSize= 1 << 24

	CommandTypeNoOperation =-1
	CommandTypeAddPeer =-2
	CommandTypeRemovePeer =-3

	DefaultNumInstallSnapshot = 1<<24
	DefaultMaxEntriesPerFile = 1 << 27
	DefaultChunkSize int64 = 1 << 24


)
type SnapshotSyncType int

const (
	Never SnapshotSyncType = 0
	EverySecond SnapshotSyncType= 1
	EveryMinute SnapshotSyncType= 2
	EveryHour SnapshotSyncType= 3
	EveryDay SnapshotSyncType= 4
	DefalutSave SnapshotSyncType= 5
	Save SnapshotSyncType= 6
	Always SnapshotSyncType= 9

	SecondsSaveSnapshot1 = 900
	ChangesSaveSnapshot1 = 1
	SecondsSaveSnapshot2 = 300
	ChangesSaveSnapshot2 = 10
	SecondsSaveSnapshot3 = 60
	ChangesSaveSnapshot3 = 10000
)



