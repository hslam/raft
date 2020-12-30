// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const lastsSize = 4
const minLatency = int64(time.Millisecond * 10)

type pipeline struct {
	node           *node
	lock           sync.Mutex
	buffer         []byte
	applyIndex     uint64
	mutex          sync.Mutex
	pending        map[uint64]*invoker
	readyEntries   []*Entry
	count          uint64
	closed         int32
	done           chan bool
	lasts          [lastsSize]int
	lastsCursor    int
	latencys       [lastsSize]int64
	latencysCursor int
	max            int
	min            int64
	lastTime       time.Time
	applying       int32
	trigger        chan bool
	readTrigger    chan bool
}

func newPipeline(n *node) *pipeline {
	p := &pipeline{
		node:        n,
		buffer:      make([]byte, 1024*64),
		pending:     make(map[uint64]*invoker),
		lastTime:    time.Now(),
		trigger:     make(chan bool, 1),
		readTrigger: make(chan bool, 1),
		done:        make(chan bool, 1),
		min:         minLatency,
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
	p.mutex.Lock()
	p.lastsCursor++
	concurrency := len(p.pending)
	p.lasts[p.lastsCursor%lastsSize] = concurrency
	var max int
	for i := 0; i < lastsSize; i++ {
		if p.lasts[i] > max {
			max = p.lasts[i]
		}
	}
	if max > p.max {
		p.max = max
	}
	p.mutex.Unlock()
	return max
}

func (p *pipeline) batch() int {
	p.mutex.Lock()
	max := p.max
	p.mutex.Unlock()
	return max
}

func (p *pipeline) updateLatency(d int64) (n int64) {
	p.mutex.Lock()
	p.latencysCursor++
	p.latencys[p.latencysCursor%lastsSize] = d
	var min int64 = minLatency
	for i := 0; i < lastsSize; i++ {
		if p.latencys[i] > 0 && p.latencys[i] < min {
			min = p.latencys[i]
		}
	}
	//logger.Tracef("pipeline.updateLatency %v,%d", p.latencys, min)
	p.min = min
	p.mutex.Unlock()
	return min
}

func (p *pipeline) minLatency() time.Duration {
	p.mutex.Lock()
	min := p.min
	p.mutex.Unlock()
	//logger.Tracef("pipeline.minLatency%d", min)
	return time.Duration(min)
}

func (p *pipeline) sleepTime() (d time.Duration) {
	if p.batch() < 1 {
		d = time.Second
	} else {
		d = p.minLatency()
	}
	return
}

func (p *pipeline) write(i *invoker) {
	p.lock.Lock()
	i.index = atomic.LoadUint64(&p.node.nextIndex)
	p.mutex.Lock()
	p.pending[i.index] = i
	p.mutex.Unlock()
	concurrency := p.concurrency()
	//logger.Tracef("pipeline.write concurrency-%d", concurrency)
	var data []byte
	if i.Command.Type() >= 0 {
		b, err := p.node.codec.Marshal(p.buffer, i.Command)
		if err != nil {
			p.lock.Unlock()
			i.Error = err
			i.done()
			return
		}
		data = make([]byte, len(b))
		copy(data, b)
	} else {
		b, err := p.node.raftCodec.Marshal(p.buffer, i.Command)
		if err != nil {
			p.lock.Unlock()
			i.Error = err
			i.done()
			return
		}
		data = make([]byte, len(b))
		copy(data, b)
	}
	entry := p.node.log.getEmtyEntry()
	entry.Index = i.index
	entry.Term = p.node.currentTerm.Load()
	entry.Command = data
	entry.CommandType = i.Command.Type()
	p.readyEntries = append(p.readyEntries, entry)
	if len(p.readyEntries) >= concurrency {
		start := time.Now().UnixNano()
		p.node.log.appendEntries(p.readyEntries)
		p.readyEntries = p.readyEntries[:0]
		go func(d int64) {
			p.updateLatency(d)
			p.lastTime = time.Now()
			p.node.check()
		}(time.Now().UnixNano() - start)
	}
	atomic.AddUint64(&p.node.nextIndex, 1)
	p.lock.Unlock()
	select {
	case p.trigger <- true:
	default:
	}
	select {
	case p.readTrigger <- true:
	default:
	}
}

func (p *pipeline) run() {
	for {
		p.lock.Lock()
		if len(p.readyEntries) > 0 {
			p.node.log.appendEntries(p.readyEntries)
			p.readyEntries = p.readyEntries[:0]
			go p.node.check()
		}
		p.lock.Unlock()
		if p.lastTime.Add(p.minLatency() * 10).Before(time.Now()) {
			p.lastTime = time.Now()
			p.mutex.Lock()
			p.max /= 2
			p.mutex.Unlock()
		}
		timer := time.NewTimer(p.sleepTime())
		runtime.Gosched()
		select {
		case <-timer.C:
		case <-p.trigger:
			timer.Stop()
			time.Sleep(p.sleepTime())
		case <-p.done:
			timer.Stop()
			return
		}
	}
}

func (p *pipeline) read() {
	for {
	loop:
		runtime.Gosched()
		time.Sleep(p.minLatency() / 10)
		p.apply()
		if atomic.LoadUint64(&p.node.nextIndex)-1 > p.node.stateMachine.lastApplied {
			goto loop
		}
		runtime.Gosched()
		select {
		case <-p.readTrigger:
			goto loop
		case <-p.done:
			return
		}
	}
}

func (p *pipeline) apply() {
	if !atomic.CompareAndSwapInt32(&p.applying, 0, 1) {
		return
	}
	if p.applyIndex-1 > p.node.stateMachine.lastApplied && p.node.commitIndex.ID() > p.node.stateMachine.lastApplied {
		//logger.Tracef("pipeline.read commitIndex-%d", p.node.commitIndex)
		p.node.log.applyCommitedEnd(p.applyIndex - 1)
	}
	var err error
	for p.applyIndex <= p.node.commitIndex.ID() {
		p.mutex.Lock()
		var i *invoker
		if in, ok := p.pending[p.applyIndex]; ok {
			i = in
		} else {
			p.mutex.Unlock()
			//Traceln("pipeline.read sleep")
			runtime.Gosched()
			time.Sleep(p.minLatency() / 100)
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
	atomic.StoreInt32(&p.applying, 0)
}

func (p *pipeline) Stop() {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return
	}
	close(p.done)
	close(p.trigger)
	close(p.readTrigger)
}
