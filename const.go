// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"time"
)

const (
	defaultStartWait              = 3 * 1000 * time.Millisecond
	defaultNodeTick               = time.Millisecond * 100
	defaultUpdateTick             = time.Millisecond * 100
	defaultNodeTracePrintTick     = 1000 * time.Millisecond
	defaultCompactionTick         = 60 * 1000 * time.Millisecond
	defaultHeartbeatTick          = 100 * time.Millisecond
	defaultDetectTick             = 100 * time.Millisecond
	defaultKeepAliveTick          = 10000 * time.Millisecond
	defaultCheckLogTick           = time.Second
	defaultElectionTimeout        = 1000 * time.Millisecond
	defaultHearbeatTimeout        = 100 * time.Millisecond
	defaultRequestVoteTimeout     = 1000 * time.Millisecond
	defaultAppendEntriesTimeout   = 10 * 1000 * time.Millisecond
	defaultInstallSnapshotTimeout = 60 * 1000 * time.Millisecond
	defaultQueryLeaderTimeout     = 60 * 1000 * time.Millisecond
	defaultAddPeerTimeout         = 60 * 1000 * time.Millisecond
	defaultRemovePeerTimeout      = 60 * 1000 * time.Millisecond
	defaultCommandTimeout         = 60 * 1000 * time.Millisecond
	defaultMaxConcurrencyRead     = 1024 * 1024
	defaultMaxConcurrency         = 1024 * 1024
	defaultMaxBatch               = 1024 * 1024
	defaultTarTick                = time.Hour
	defaultMaxDelay               = time.Millisecond
	defaultDataDir                = "default.raft"
	defaultConfig                 = "config"
	defaultLog                    = "log"
	defaultIndex                  = "index"
	defaultTerm                   = "term"
	defaultVoteFor                = "votefor"
	defaultSnapshot               = "snapshot"
	defaultLastIncludedIndex      = "lastincludedindex"
	defaultLastIncludedTerm       = "lastincludedterm"
	defaultLastTarIndex           = "lasttarindex"
	defaultMd5                    = "md5"
	defaultTar                    = "tar"
	defaultTarGz                  = "tar.gz"
	defaultTmp                    = ".tmp"
	defaultFlush                  = ".flush"
	defaultReadFileBufferSize     = 1 << 24

	commandTypeNoOperation           = -1
	commandTypeAddPeer               = -2
	commandTypeRemovePeer            = -3
	commandTypeReconfiguration       = -4
	defaultNumInstallSnapshot        = 1 << 24
	defaultMaxEntriesPerFile         = 1 << 27
	defaultChunkSize           int64 = 1 << 24
)

// SnapshotPolicy represents a snapshot policy type.
type SnapshotPolicy int

const (
	// Never is a SnapshotPolicy that will never sync the snapshot to the disk.
	Never SnapshotPolicy = 0
	// EverySecond is a SnapshotPolicy that will sync the snapshot to the disk every second.
	EverySecond SnapshotPolicy = 1
	// EveryMinute is a SnapshotPolicy that will sync the snapshot to the disk every minute.
	EveryMinute SnapshotPolicy = 2
	// EveryHour is a SnapshotPolicy that will sync the snapshot to the disk every hour.
	EveryHour SnapshotPolicy = 3
	// EveryDay is a SnapshotPolicy that will sync the snapshot to the disk every day.
	EveryDay SnapshotPolicy = 4
	// DefalutSync is a defalut SnapshotPolicy.
	DefalutSync SnapshotPolicy = 5
	// CustomSync is a custom SnapshotPolicy.
	CustomSync SnapshotPolicy = 6
	// Always is a SnapshotPolicy that will sync the snapshot to the disk every command.
	Always SnapshotPolicy = 9
)

const (
	secondsSaveSnapshot1 = 900
	changesSaveSnapshot1 = 1
	secondsSaveSnapshot2 = 300
	changesSaveSnapshot2 = 10
	secondsSaveSnapshot3 = 60
	changesSaveSnapshot3 = 10000
)