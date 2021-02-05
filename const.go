// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"time"
)

const (
	defaultStartWait                    = 3 * 1000 * time.Millisecond
	defaultNodeTick                     = time.Millisecond * 100
	defaultUpdateTick                   = time.Millisecond * 100
	defaultNodeTracePrintTick           = 1000 * time.Millisecond
	defaultHeartbeatTick                = 100 * time.Millisecond
	defaultDetectTick                   = 100 * time.Millisecond
	defaultKeepAliveTick                = 10 * 1000 * time.Millisecond
	defaultCheckLogTick                 = time.Second
	defaultElectionTimeout              = 1000 * time.Millisecond
	defaultHearbeatTimeout              = 100 * time.Millisecond
	defaultRequestVoteTimeout           = 1000 * time.Millisecond
	defaultAppendEntriesTimeout         = 60 * 1000 * time.Millisecond
	defaultInstallSnapshotTimeout       = 60 * 1000 * time.Millisecond
	defaultGetLeaderTimeout             = 60 * 1000 * time.Millisecond
	defaultAddMemberTimeout             = 60 * 1000 * time.Millisecond
	defaultRemoveMemberTimeout          = 60 * 1000 * time.Millisecond
	defaultSetMetaTimeout               = 60 * 1000 * time.Millisecond
	defaultGetMetaTimeout               = 60 * 1000 * time.Millisecond
	defaultCommandTimeout               = 60 * 1000 * time.Millisecond
	defaultMatchTimeout                 = 60 * 1000 * time.Millisecond
	defaultMaxBatch                     = 32 * 1024
	defaultMaxCache                     = 32 * 1024
	defaultTarTick                      = time.Second
	defaultMaxDelay                     = time.Millisecond
	defaultDataDir                      = "default.raft"
	defaultConfig                       = "config"
	defaultCommitIndex                  = "commitindex"
	defaultSnapshot                     = "snapshot"
	defaultLastIncludedIndex            = "lastincludedindex"
	defaultLastIncludedTerm             = "lastincludedterm"
	defaultLastTarIndex                 = "lasttarindex"
	defaultTar                          = "tar"
	defaultTarGz                        = "tar.gz"
	defaultTmp                          = ".tmp"
	defaultFlush                        = ".flush"
	defaultChunkSize              int64 = 1 << 24
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
)

const (
	secondsSaveSnapshot1 = 900
	changesSaveSnapshot1 = 1
	secondsSaveSnapshot2 = 300
	changesSaveSnapshot2 = 10
	secondsSaveSnapshot3 = 60
	changesSaveSnapshot3 = 10000
)
