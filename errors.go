package raft

import (
	"errors"
)

var (
	ErrNotLeader=errors.New("this node is not Leader")
	ErrNotRunning=errors.New("this node do not running")
	ErrLeaderIsNotReady=errors.New("Leader is not ready")


	//AppendEntries
	ErrAppendEntriesFailed=errors.New("AppendEntries failed")
	ErrAppendEntriesTimeout=errors.New("AppendEntries timeout")


	//Command
	ErrCommandNil=errors.New("Command can not be nil")
	ErrCommandTimeout=errors.New("command timeout")
	ErrCommandNotRegistered=errors.New("Command is not registered")
	ErrCommandTypeExisted=errors.New("CommandType is existed")
	ErrCommandTypeMinus=errors.New("CommandType must be >=0")


	//SnapshotCodec
	ErrSnapshotCodecNil=errors.New("SnapshotCodec can not be nil")
)
