package raft

import (
	"sync"
	"errors"
)

type PersistentString struct {
	mu 								sync.RWMutex
	node 							*Node
	value							string
	name 							string
}
func newPersistentString(node *Node,name string) *PersistentString {
	p:=&PersistentString{
		node:node,
		name:name,
	}
	p.load()
	return p
}
func (p *PersistentString) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value=""
	p.save()
}
func (p *PersistentString) Set(v string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value=v
	p.save()
}

func (p *PersistentString) String()string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *PersistentString) save() {
	p.node.storage.SafeOverWrite(p.name,[]byte(p.value))
}

func (p *PersistentString) load() error {
	if !p.node.storage.Exists(p.name){
		return errors.New(p.name+" file is not existed")
	}
	b, err := p.node.storage.Load(p.name)
	if err != nil {
		return err
	}
	p.value = string(b)
	return nil
}