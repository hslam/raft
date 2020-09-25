// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"fmt"
	"github.com/hslam/code"
	"sync"
	"time"
)

type persistentUint64 struct {
	mu            sync.RWMutex
	node          *Node
	value         uint64
	name          string
	ticker        *time.Ticker
	lastSaveValue uint64
	deferSave     bool
}

func newPersistentUint64(node *Node, name string, tick time.Duration) *persistentUint64 {
	p := &persistentUint64{
		node: node,
		name: name,
	}
	if tick > 0 {
		p.deferSave = true
		p.ticker = time.NewTicker(tick)
		go p.run()
	}
	p.load()
	return p
}
func (p *persistentUint64) Incre() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value += 1
	if !p.deferSave {
		p.save()
	}
	return p.value
}
func (p *persistentUint64) Set(t uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = t
	if !p.deferSave {
		p.save()
	}
}
func (p *persistentUint64) Id() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *persistentUint64) save() {
	buf := make([]byte, 8)
	code.EncodeUint64(buf, p.value)
	p.node.storage.OverWrite(p.name, buf)
}

func (p *persistentUint64) load() error {
	if !p.node.storage.Exists(p.name) {
		p.value = 0
		return errors.New(p.name + " file is not existed")
	}
	b, err := p.node.storage.Load(p.name)
	if err != nil {
		return err
	}
	if len(b) != 8 {
		return fmt.Errorf("length %d", len(b))
	}
	code.DecodeUint64(b, &p.value)
	return nil
}

func (p *persistentUint64) run() {
	for range p.ticker.C {
		if p.lastSaveValue != p.value {
			func() {
				p.mu.Lock()
				defer p.mu.Unlock()
				p.save()
			}()
			p.lastSaveValue = p.value
		}
	}
}
func (p *persistentUint64) Stop() {
	p.ticker.Stop()
	p.ticker = nil
}
