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
const latency = int64(time.Millisecond * 10)

type pipe struct {
	node           *node
	lock           sync.Mutex
	buffer         []byte
	applyIndex     uint64
	mutex          sync.Mutex
	pending        map[uint64]*invoker
	readyEntries   []*Entry
	appendEntries  chan []*Entry
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

func newPipe(n *node) *pipe {
	p := &pipe{
		node:          n,
		buffer:        make([]byte, 1024*64),
		pending:       make(map[uint64]*invoker),
		appendEntries: make(chan []*Entry, 1024),
		lastTime:      time.Now(),
		trigger:       make(chan bool),
		readTrigger:   make(chan bool, 1),
		done:          make(chan bool, 1),
		min:           latency,
	}
	go p.append()
	go p.run()
	go p.read()
	return p
}

func (p *pipe) init(lastLogIndex uint64) {
	//logger.Tracef("pipe.init %d", lastLogIndex)
	p.applyIndex = lastLogIndex + 1
}

func (p *pipe) concurrency() (n int) {
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

func (p *pipe) batch() int {
	p.mutex.Lock()
	max := p.max
	p.mutex.Unlock()
	return max
}

func (p *pipe) updateLatency(d int64) (n int64) {
	p.mutex.Lock()
	p.latencysCursor++
	p.latencys[p.latencysCursor%lastsSize] = d
	var min int64 = latency
	for i := 0; i < lastsSize; i++ {
		if p.latencys[i] > 0 && p.latencys[i] < min {
			min = p.latencys[i]
		}
	}
	//logger.Tracef("pipe.updateLatency %v,%d", p.latencys, min)
	p.min = min
	p.mutex.Unlock()
	return min
}

func (p *pipe) minLatency() time.Duration {
	p.mutex.Lock()
	min := p.min
	p.mutex.Unlock()
	//logger.Tracef("pipe.minLatency%d", min)
	return time.Duration(min)
}

func (p *pipe) sleepTime() (d time.Duration) {
	if p.batch() < 1 {
		d = time.Second
	} else {
		d = p.minLatency()
	}
	return
}

func (p *pipe) write(i *invoker) {
	p.lock.Lock()
	i.index = atomic.LoadUint64(&p.node.nextIndex)
	p.mutex.Lock()
	p.pending[i.index] = i
	p.mutex.Unlock()
	concurrency := p.concurrency()
	//logger.Tracef("pipe.write concurrency-%d", concurrency)
	var data []byte
	var codec Codec
	if i.Command.Type() > 0 {
		codec = p.node.codec
	} else {
		codec = p.node.raftCodec
	}
	b, err := codec.Marshal(p.buffer, i.Command)
	if err != nil {
		p.lock.Unlock()
		i.Error = err
		i.done()
		return
	}
	data = make([]byte, len(b))
	copy(data, b)
	entry := p.node.log.getEmtyEntry()
	entry.Index = i.index
	entry.Term = p.node.currentTerm.Load()
	entry.Command = data
	entry.CommandType = i.Command.Type()
	p.readyEntries = append(p.readyEntries, entry)
	if len(p.readyEntries) >= concurrency {
		entries := p.readyEntries
		p.appendEntries <- entries
		p.readyEntries = []*Entry{}
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

func (p *pipe) append() {
	for {
		runtime.Gosched()
		select {
		case entries, ok := <-p.appendEntries:
			if ok && len(entries) > 0 {
				//logger.Tracef("pipe.write concurrency-%d,entries-%d,sleep-%v", p.batch(), len(entries), p.sleepTime())
				start := time.Now().UnixNano()
				p.node.log.appendEntries(entries)
				go func(d int64) {
					p.updateLatency(d)
					p.lastTime = time.Now()
					p.node.check()
				}(time.Now().UnixNano() - start)
			}
		case <-p.done:
			return
		}
	}
}

func (p *pipe) run() {
	for {
		p.lock.Lock()
		if len(p.readyEntries) > 0 {
			entries := p.readyEntries
			p.appendEntries <- entries
			p.readyEntries = []*Entry{}
			go p.node.check()
		}
		p.lock.Unlock()
		timer := time.NewTimer(p.sleepTime())
		runtime.Gosched()
		select {
		case <-timer.C:
		case <-p.trigger:
			timer.Stop()
			time.Sleep(p.sleepTime() / 500 * time.Duration(p.concurrency()))
		case <-p.done:
			timer.Stop()
			return
		}
	}
}

func (p *pipe) read() {
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

func (p *pipe) apply() {
	if !atomic.CompareAndSwapInt32(&p.applying, 0, 1) {
		return
	}
	if p.applyIndex-1 > p.node.stateMachine.lastApplied && p.node.commitIndex.ID() > p.node.stateMachine.lastApplied {
		//logger.Tracef("pipe.read commitIndex-%d", p.node.commitIndex)
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
			//Traceln("pipe.read sleep")
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

func (p *pipe) Stop() {
	if atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		close(p.done)
	}
}
