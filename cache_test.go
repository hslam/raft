// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"testing"
)

func TestCache(t *testing.T) {
	size := 3
	c := newCache(size)
	if c.length() != 0 {
		t.Error()
	}
	c.dequeue()
	if entries := c.CopyAfter(0, size); len(entries) != 0 {
		t.Error()
	}
	index := uint64(1000000)
	for i := uint64(0); i < uint64(size); i++ {
		index++
		c.AppendEntries([]*Entry{{Index: index}})
	}
	if entries := c.CopyAfter(index-10000, size); len(entries) != 0 {
		t.Error()
	}
	if entries := c.CopyAfter(index+10000, size); len(entries) != 0 {
		t.Error()
	}
	for i := uint64(0); i < 100; i++ {
		index++
		c.AppendEntries([]*Entry{{Index: index}})
		if c.firstIndex() > c.lastIndex() {
			t.Error()
		}
		entries := c.CopyAfter(index-uint64(size)+1, size)
		if len(entries) != size {
			t.Error()
		}
	}
	c.Reset()
	if c.length() != 0 {
		t.Error()
	}
}

func TestReset(t *testing.T) {
	size := 4096
	c := newCache(4096)
	index := uint64(1000000)
	for i := uint64(0); i < uint64(size); i++ {
		index++
		c.AppendEntries([]*Entry{{Index: index}})
	}
	c.Reset()
	if c.length() != 0 {
		t.Error()
	}
	for i := range c.entries {
		if c.entries[i] != nil {
			t.Error()
		}
	}
}
