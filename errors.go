package raft

import (
	"errors"
)

var (
	ErrNotLeader=errors.New("this node is not Leader")
	ErrCommandTimeout=errors.New("command timeout")
	ErrCommandNotRegistered=errors.New("Command is not registered")
	ErrCommandTypeExisted=errors.New("CommandType is existed")
)
