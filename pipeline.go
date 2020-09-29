// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

const lastsSize = 4
const minLatency = int64(time.Millisecond * 10)

type pipeline struct {
	node           *node
	wMutex         sync.Mutex
	buffer         []byte
	applyIndex     uint64
	mutex          sync.Mutex
	pending        map[uint64]*invoker
	readyEntries   []*Entry
	bMutex         sync.Mutex
	count          uint64
	done           chan bool
	lasts          [lastsSize]int
	lastsCursor    int
	latencys       [lastsSize]int64
	latencysCursor int
	max            int
	min            int64
	lastTime       time.Time
	trigger        chan bool
}

func newPipeline(n *node) *pipeline {
	p := &pipeline{
		node:     n,
		buffer:   make([]byte, 1024*64),
		pending:  make(map[uint64]*invoker),
		lastTime: time.Now(),
		trigger:  make(chan bool, 1024),
		done:     make(chan bool, 1),
		min:      minLatency,
	}
	go p.run()
	go p.read()
	return p
}

func (p *pipeline) init(lastLogIndex uint64) {
	//logger.Tracef("pipeline.init %d", lastLogIndex)
	p.applyIndex = lastLogIndex + 1
}
func (p *pipeline) concurrency() (n int) {
	p.lastsCursor++
	p.mutex.Lock()
	concurrency := len(p.pending)
	p.lasts[p.lastsCursor%lastsSize] = concurrency
	p.mutex.Unlock()
	var max int
	for i := 0; i < lastsSize; i++ {
		if p.lasts[i] > max {
			max = p.lasts[i]
		}
	}
	if max > p.max {
		p.max = max
	}
	return p.max
}
func (p *pipeline) updateLatency(d int64) (n int64) {
	p.latencysCursor++
	p.mutex.Lock()
	p.latencys[p.latencysCursor%lastsSize] = d
	p.mutex.Unlock()
	var min int64 = minLatency
	for i := 0; i < lastsSize; i++ {
		if p.latencys[i] > 0 && p.latencys[i] < min {
			min = p.latencys[i]
		}
	}
	//logger.Tracef("pipeline.updateLatency %v,%d", p.latencys, min)
	p.min = min
	return p.min
}
func (p *pipeline) minLatency() int64 {
	p.mutex.Lock()
	min := p.min
	p.mutex.Unlock()
	//logger.Tracef("pipeline.minLatency%d", min)
	return min
}

func (p *pipeline) sleepTime() (d time.Duration) {
	if p.concurrency() < 1 {
		d = time.Second
	} else {
		d = time.Duration(p.minLatency())
	}
	return
}

func (p *pipeline) Update() bool {
	if p.concurrency() > 0 {
		return true
	}
	return false
}

func (p *pipeline) write(i *invoker) {
	p.wMutex.Lock()
	defer p.wMutex.Unlock()
	nextIndex := atomic.AddUint64(&p.node.nextIndex, 1)
	i.index = nextIndex - 1
	p.mutex.Lock()
	p.pending[i.index] = i
	p.mutex.Unlock()
	concurrency := p.concurrency()
	//logger.Tracef("pipeline.write concurrency-%d", concurrency)
	var data []byte
	if i.Command.Type() >= 0 {
		b, _ := p.node.codec.Marshal(p.buffer, i.Command)
		data = make([]byte, len(b))
		copy(data, b)
	} else {
		b, _ := p.node.raftCodec.Marshal(p.buffer, i.Command)
		data = make([]byte, len(b))
		copy(data, b)
	}
	entry := p.node.log.getEmtyEntry()
	entry.Index = i.index
	entry.Term = p.node.currentTerm.ID()
	entry.Command = data
	entry.CommandType = i.Command.Type()
	p.bMutex.Lock()
	p.readyEntries = append(p.readyEntries, entry)
	if len(p.readyEntries) >= concurrency {
		start := time.Now().UnixNano()
		p.node.log.ticker(p.readyEntries)
		p.readyEntries = p.readyEntries[:0]
		p.updateLatency(time.Now().UnixNano() - start)
	}
	p.bMutex.Unlock()
	select {
	case p.trigger <- true:
	default:
	}
}
func (p *pipeline) run() {
	for {
		p.bMutex.Lock()
		if len(p.readyEntries) > 0 {
			p.node.log.ticker(p.readyEntries)
			p.readyEntries = p.readyEntries[:0]
		}
		p.bMutex.Unlock()
		var d time.Duration
		if p.lastTime.Add(time.Millisecond * 10).Before(time.Now()) {
			p.lastTime = time.Now()
			p.mutex.Lock()
			p.max = 0
			p.mutex.Unlock()
		}
		d = p.sleepTime()
		select {
		case <-time.After(d):
		case <-p.trigger:
			d = p.sleepTime()
			time.Sleep(time.Duration(d))
		case <-p.done:
			return
		}
	}
}

func (p *pipeline) read() {
	var err error
	for err == nil {
		if p.node.isLeader() {
			if p.applyIndex-1 > p.node.stateMachine.lastApplied && p.node.commitIndex > p.node.stateMachine.lastApplied {
				//logger.Tracef("pipeline.read commitIndex-%d", p.node.commitIndex)
				p.node.log.applyCommitedEnd(p.applyIndex - 1)
			}
			for p.applyIndex <= p.node.commitIndex {
				p.mutex.Lock()
				var i *invoker
				if in, ok := p.pending[p.applyIndex]; ok {
					i = in
				} else {
					p.mutex.Unlock()
					//Traceln("pipeline.read sleep")
					time.Sleep(time.Microsecond * 100)
					continue
				}
				delete(p.pending, p.applyIndex)
				p.mutex.Unlock()
				i.Reply, i.Error, err = p.node.stateMachine.Apply(i.index, i.Command)
				if err != nil {
					continue
				}
				i.done()
				p.mutex.Lock()
				p.applyIndex++
				p.mutex.Unlock()
			}
			d := p.sleepTime() / 50
			time.Sleep(d)
			//logger.Tracef("pipeline.read sleepTime-%v", d)
		} else {
			time.Sleep(p.sleepTime())
		}
	}
}
