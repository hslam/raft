package raft

import (
	"sync"
	"errors"
	"fmt"
	"time"
)

type PersistentUint64 struct {
	mu 								sync.RWMutex
	node 							*Node
	value							uint64
	name 							string
	ticker 							*time.Ticker
	lastSaveValue 					uint64
	deferSave  						bool
}
func newPersistentUint64(node *Node,name string,tick time.Duration) *PersistentUint64 {
	p:=&PersistentUint64{
		node:node,
		name:name,
	}
	if tick>0{
		p.deferSave=true
		p.ticker=time.NewTicker(tick)
		go p.run()
	}
	p.load()
	return p
}
func (p *PersistentUint64) Incre()uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value+=1
	if !p.deferSave{
		p.save()
	}
	return p.value
}
func (p *PersistentUint64) Set(t uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value=t
	if !p.deferSave{
		p.save()
	}
}
func (p *PersistentUint64) Id()uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *PersistentUint64) save() {
	p.node.storage.OverWrite(p.name,uint64ToBytes(p.value))
}

func (p *PersistentUint64) load() error {
	if !p.node.storage.Exists(p.name){
		p.value=0
		return errors.New(p.name+" file is not existed")
	}
	b, err := p.node.storage.Load(p.name)
	if err != nil {
		return err
	}
	if len(b)!=8{
		return fmt.Errorf("length %d",len(b))
	}
	p.value = bytesToUint64(b)
	return nil
}

func (p *PersistentUint64)run()  {
	for range p.ticker.C{
		if p.lastSaveValue!=p.value{
			func(){
				p.mu.Lock()
				defer p.mu.Unlock()
				p.save()
			}()
			p.lastSaveValue=p.value
		}
	}
}
func (p *PersistentUint64)Stop()  {
	p.ticker.Stop()
	p.ticker=nil
}