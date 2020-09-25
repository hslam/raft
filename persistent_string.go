// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"fmt"
	"sync"
)

type persistentString struct {
	mu    sync.RWMutex
	node  *Node
	value string
	name  string
}

func newPersistentString(node *Node, name string) *persistentString {
	p := &persistentString{
		node: node,
		name: name,
	}
	p.load()
	return p
}
func (p *persistentString) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = ""
	p.save()
}
func (p *persistentString) Set(v string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = v
	p.save()
}

func (p *persistentString) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *persistentString) save() {
	p.node.storage.OverWrite(p.name, []byte(p.value))
}

func (p *persistentString) load() error {
	if !p.node.storage.Exists(p.name) {
		return errors.New(p.name + " file is not existed")
	}
	b, err := p.node.storage.Load(p.name)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return fmt.Errorf("length %d", len(b))
	}
	p.value = string(b)
	return nil
}
