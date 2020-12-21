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
	node  *node
	value string
	name  string
}

func newPersistentString(n *node, name string) *persistentString {
	p := &persistentString{
		node: n,
		name: name,
	}
	p.load()
	return p
}
func (p *persistentString) Reset() {
	p.mu.Lock()
	p.value = ""
	p.save()
	p.mu.Unlock()
}
func (p *persistentString) Set(v string) {
	p.mu.Lock()
	p.value = v
	p.save()
	p.mu.Unlock()
}

func (p *persistentString) String() string {
	p.mu.RLock()
	value := p.value
	p.mu.RUnlock()
	return value
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
