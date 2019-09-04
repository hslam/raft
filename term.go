package raft

import (
	"sync"
	"errors"
)

type Term struct {
	mu 								sync.RWMutex
	node 							*Node
	id								uint64
}
func newTerm(node *Node) *Term {
	term:=&Term{
		node:node,
	}
	term.loadTerm()
	return term
}
func (term *Term) Incre()uint64 {
	term.mu.Lock()
	defer term.mu.Unlock()
	term.id+=1
	term.saveTerm()
	return term.id
}
func (term *Term) Set(t uint64) {
	term.mu.Lock()
	defer term.mu.Unlock()
	term.id=t
	term.saveTerm()
}

func (term *Term) Id()uint64 {
	term.mu.RLock()
	defer term.mu.RUnlock()
	return term.id
}

func (term *Term) saveTerm() {
	term.node.storage.SafeOverWrite(DefaultTerm,uint64ToBytes(term.id))
}

func (term *Term) loadTerm() error {
	if !term.node.storage.Exists(DefaultTerm){
		return errors.New("term is not existed")
	}
	b, err := term.node.storage.Load(DefaultTerm)
	term.id = bytesToUint64(b)
	if err != nil {
		return nil
	}
	return nil
}