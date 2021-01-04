// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
)

var (
	// ErrNotLeader is returned when this node is not Leader.
	ErrNotLeader = errors.New("this node is not Leader")
	// ErrNotRunning is returned when this node do not running.
	ErrNotRunning = errors.New("this node do not running")
	// ErrLeaderIsNotReady is returned when the leader is not ready.
	ErrLeaderIsNotReady = errors.New("Leader is not ready")

	// ErrAppendEntriesFailed is returned when append entries failed.
	ErrAppendEntriesFailed = errors.New("AppendEntries failed")
	// ErrAppendEntriesTimeout is returned when append entries timeout.
	ErrAppendEntriesTimeout = errors.New("AppendEntries timeout")

	// ErrCommandNil is returned when the command is nil.
	ErrCommandNil = errors.New("Command can not be nil")
	// ErrCommandTimeout is returned when exec the command timeout.
	ErrCommandTimeout = errors.New("command timeout")
	// ErrCommandNotRegistered is returned when the command is not registered.
	ErrCommandNotRegistered = errors.New("Command is not registered")
	// ErrCommandTypeExisted is returned when the command type is existed.
	ErrCommandTypeExisted = errors.New("CommandType is existed")
	// ErrCommandType is returned when the command type = 0.
	ErrCommandType = errors.New("CommandType must be > 0")

	// ErrSnapshotCodecNil is returned when the snapshotCodec is nil.
	ErrSnapshotCodecNil = errors.New("SnapshotCodec can not be nil")
)
