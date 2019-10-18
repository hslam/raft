package raft

import (
	"sync"
	"errors"
	"fmt"
)

type PersistentUint64 struct {
	mu 								sync.RWMutex
	node 							*Node
	value							uint64
	name 							string
}
func newPersistentUint64(node *Node,name string) *PersistentUint64 {
	p:=&PersistentUint64{
		node:node,
		name:name,
	}
	p.load()
	return p
}
func (p *PersistentUint64) Incre()uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value+=1
	p.save()
	return p.value
}
func (p *PersistentUint64) Set(t uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value=t
	p.save()
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