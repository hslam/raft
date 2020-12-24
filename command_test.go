// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"testing"
)

func TestCommand(t *testing.T) {
	commands := &commands{types: make(map[int32]*sync.Pool)}
	if err := commands.register(&testCommand{}); err != nil {
		t.Error()
	}
	cmd := &testCommand{}
	if err := commands.register(cmd); err == nil {
		t.Error()
	}
	if cp := commands.clone(cmd.Type()); cp == nil {
		t.Error()
	}
	if cp := commands.clone(cmd.Type() + 1); cp != nil {
		t.Error()
	}
}
