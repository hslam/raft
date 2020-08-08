package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

const lastsSize = 4
const minLatency = int64(time.Millisecond * 10)

type pipeline struct {
	node           *Node
	wMutex         sync.Mutex
	buffer         []byte
	applyIndex     uint64
	mutex          sync.Mutex
	pending        map[uint64]*Invoker
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
	triggerCnt     int64
	trigger        chan bool
}

func newPipeline(node *Node) *pipeline {
	p := &pipeline{
		node:     node,
		buffer:   make([]byte, 1024*64),
		pending:  make(map[uint64]*Invoker),
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
	//Tracef("pipeline.init %d", lastLogIndex)
	p.applyIndex = lastLogIndex + 1
}
func (p *pipeline) concurrency() (n int) {
	p.lastsCursor += 1
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
	p.latencysCursor += 1
	p.mutex.Lock()
	p.latencys[p.latencysCursor%lastsSize] = d
	p.mutex.Unlock()
	var min int64 = minLatency
	for i := 0; i < lastsSize; i++ {
		if p.latencys[i] > 0 && p.latencys[i] < min {
			min = p.latencys[i]
		}
	}
	//Tracef("pipeline.updateLatency %v,%d", p.latencys, min)
	p.min = min
	return p.min
}
func (p *pipeline) minLatency() int64 {
	p.mutex.Lock()
	min := p.min
	p.mutex.Unlock()
	return min
}
func (p *pipeline) write(invoker *Invoker) {
	p.wMutex.Lock()
	defer p.wMutex.Unlock()
	nextIndex := atomic.AddUint64(&p.node.nextIndex, 1)
	invoker.index = nextIndex - 1
	p.mutex.Lock()
	p.pending[invoker.index] = invoker
	p.mutex.Unlock()
	concurrency := p.concurrency()
	//Tracef("pipeline.write concurrency-%d", concurrency)
	var data []byte
	if invoker.Command.Type() >= 0 {
		b, _ := p.node.codec.Marshal(p.buffer, invoker.Command)
		data = make([]byte, len(b))
		copy(data, b)
	} else {
		b, _ := p.node.raftCodec.Marshal(p.buffer, invoker.Command)
		data = make([]byte, len(b))
		copy(data, b)
	}
	entry := p.node.log.getEmtyEntry()
	entry.Index = invoker.index
	entry.Term = p.node.currentTerm.Id()
	entry.Command = data
	entry.CommandType = invoker.Command.Type()
	entry.CommandId = invoker.Command.UniqueID()
	p.bMutex.Lock()
	p.readyEntries = append(p.readyEntries, entry)
	if len(p.readyEntries) >= concurrency {
		start := time.Now().UnixNano()
		p.node.log.ticker(p.readyEntries)
		p.readyEntries = p.readyEntries[:0]
		p.updateLatency(time.Now().UnixNano() - start)
	}
	p.bMutex.Unlock()
	if atomic.LoadInt64(&p.triggerCnt) < 1 {
		atomic.AddInt64(&p.triggerCnt, 1)
		p.trigger <- true
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
		if p.lastTime.Add(time.Microsecond * 10).Before(time.Now()) {
			p.lastTime = time.Now()
			p.mutex.Lock()
			p.max = 0
			p.mutex.Unlock()
		}
		if p.concurrency() < 1 {
			d = time.Second
		} else {
			d = time.Duration(p.minLatency())
		}
		select {
		case <-time.After(d):
		case <-p.trigger:
			if p.concurrency() < 1 {
				d = time.Second
			} else {
				d = time.Duration(p.minLatency())
			}
			time.Sleep(time.Duration(d))
			atomic.StoreInt64(&p.triggerCnt, 0)
		case <-p.done:
			return
		}
	}
}

func (p *pipeline) read() {
	var err error
	for err == nil {
		if p.applyIndex-1 > p.node.stateMachine.lastApplied && p.node.commitIndex > p.node.stateMachine.lastApplied {
			//Tracef("pipeline.read commitIndex-%d", p.node.commitIndex)
			p.node.log.applyCommitedEnd(p.applyIndex - 1)
		}
		for p.applyIndex <= p.node.commitIndex {
			p.mutex.Lock()
			var invoker *Invoker
			if i, ok := p.pending[p.applyIndex]; ok {
				invoker = i
			} else {
				p.mutex.Unlock()
				time.Sleep(time.Microsecond * 100)
				continue
			}
			delete(p.pending, p.applyIndex)
			p.mutex.Unlock()
			invoker.Reply, invoker.Error, err = p.node.stateMachine.Apply(invoker.index, invoker.Command)
			if err != nil {
				continue
			}
			invoker.done()
			p.mutex.Lock()
			p.applyIndex++
			p.mutex.Unlock()
		}
		time.Sleep(time.Microsecond * 100)
	}
}
