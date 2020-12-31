// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"fmt"
	"github.com/hslam/code"
	"sync"
	"sync/atomic"
	"time"
)

type persistentUint64 struct {
	mu            sync.RWMutex
	node          *node
	value         uint64
	name          string
	offset        uint64
	ticker        *time.Ticker
	lastSaveValue uint64
	deferSave     bool
	buf           []byte
	done          chan struct{}
	closed        uint32
}

func newPersistentUint64(n *node, name string, offset uint64, tick time.Duration) *persistentUint64 {
	p := &persistentUint64{
		node:   n,
		name:   name,
		offset: offset,
		buf:    make([]byte, 8),
		done:   make(chan struct{}, 1),
	}
	if tick > 0 {
		p.deferSave = true
		p.ticker = time.NewTicker(tick)
		go p.run()
	}
	p.load()
	return p
}

func (p *persistentUint64) Set(t uint64) {
	atomic.StoreUint64(&p.value, t)
	if !p.deferSave {
		p.save()
	}
}

func (p *persistentUint64) ID() uint64 {
	return atomic.LoadUint64(&p.value)
}

func (p *persistentUint64) save() {
	p.mu.Lock()
	code.EncodeUint64(p.buf, atomic.LoadUint64(&p.value))
	if p.deferSave {
		p.node.storage.SeekWriteNoSync(p.name, p.offset, p.buf)
	} else {
		p.node.storage.SeekWrite(p.name, p.offset, p.buf)
	}
	p.mu.Unlock()
}

func (p *persistentUint64) load() error {
	if !p.node.storage.Exists(p.name) {
		p.value = 0
		return errors.New(p.name + " file is not existed")
	}
	n, err := p.node.storage.SeekRead(p.name, p.offset, p.buf)
	if err != nil {
		return err
	}
	if n != 8 {
		return fmt.Errorf("length %d", n)
	}
	code.DecodeUint64(p.buf, &p.value)
	return nil
}

func (p *persistentUint64) run() {
	for {
		select {
		case <-p.ticker.C:
			value := atomic.LoadUint64(&p.value)
			if p.lastSaveValue != value {
				p.save()
				p.lastSaveValue = value
			}
		case <-p.done:
			return
		}
	}
}

func (p *persistentUint64) Stop() {
	if atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		p.ticker.Stop()
		close(p.done)
	}
}
