// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
)

type cache struct {
	lock    sync.RWMutex
	entries []*Entry
	front   uint64
	rear    uint64
	maxSize uint64
}

func newCache(size int) *cache {
	size++
	return &cache{
		entries: make([]*Entry, size),
		front:   0,
		rear:    0,
		maxSize: uint64(size),
	}
}

var emptyEntries = make([]*Entry, 1024)

func (c *cache) Reset() {
	c.lock.Lock()
	c.front = 0
	c.rear = 0
	cursor := uint64(0)
	for cursor < c.maxSize {
		n := copy(c.entries[cursor:], emptyEntries)
		cursor += uint64(n)
	}
	c.lock.Unlock()
}

func (c *cache) AppendEntries(entries []*Entry) {
	c.lock.Lock()
	for i := 0; i < len(entries); i++ {
		c.enqueue(entries[i])
	}
	c.lock.Unlock()
}

func (c *cache) CopyAfter(index uint64, max int) (entries []*Entry) {
	c.lock.RLock()
	if c.length() == 0 {
		c.lock.RUnlock()
		return
	}
	firstIndex := c.firstIndex()
	lastIndex := c.lastIndex()
	if index < firstIndex {
		c.lock.RUnlock()
		return
	}
	if index > lastIndex {
		c.lock.RUnlock()
		return
	}
	startIndex := index
	endIndex := minUint64(startIndex+uint64(max), lastIndex)
	if startIndex <= endIndex {
		entries = make([]*Entry, 0, endIndex-startIndex+1)
		for i := startIndex; i < endIndex+1; i++ {
			entry := c.entries[(c.front+i-firstIndex)%c.maxSize]
			entries = append(entries, entry)
		}
	}
	c.lock.RUnlock()
	return
}

func (c *cache) firstIndex() (index uint64) {
	return c.entries[c.front].Index
}

func (c *cache) lastIndex() (index uint64) {
	return c.entries[(c.rear-1+c.maxSize)%c.maxSize].Index
}

func (c *cache) length() uint64 {
	return (c.rear - c.front + c.maxSize) % c.maxSize
}

func (c *cache) enqueue(entry *Entry) {
	if (c.rear+1)%c.maxSize == c.front {
		c.dequeue()
	}
	c.entries[c.rear] = entry
	c.rear = (c.rear + 1) % c.maxSize
}

func (c *cache) dequeue() (entry *Entry) {
	if c.rear == c.front {
		return
	}
	entry = c.entries[c.front]
	c.entries[c.front] = nil
	c.front = (c.front + 1) % c.maxSize
	return
}
