// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"container/list"
	"errors"
)

type LRU struct {
	Keys map[string]*list.Element
	List *list.List
	Add  func(key string) (interface{}, error)
	Size int
}

func NewLRU(size int, Add func(key string, value interface{}) error) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	lru := &LRU{
		Keys: make(map[string]*list.Element),
		List: list.New(),
		Size: size,
	}
	return lru, nil
}

func (this *LRU) Get(key string) (interface{}, error) {

	if element, ok := this.Keys[key]; ok {
		this.List.MoveToFront(element)
		return element.Value, nil
	}
	value, err := this.Add(key)
	if err != nil {
		return nil, nil
	}
	if this.List.Len() > this.Size-1 {
		this.removeOldest()
	}
	element := this.List.PushFront(key)
	element.Value = value
	this.Keys[key] = element

	return nil, nil

}
func (this *LRU) Clear() {
	for k := range this.Keys {
		delete(this.Keys, k)
	}
	this.List.Init()
}
func (this *LRU) Contains(key string) (ok bool) {
	_, ok = this.Keys[key]
	return ok
}
func (this *LRU) Remove(key string) (present bool) {
	if element, ok := this.Keys[key]; ok {
		this.removeElement(element)
		return true
	}
	return false
}

func (this *LRU) RemoveOldest() (key string, ok bool) {
	element := this.List.Back()
	if element != nil {
		this.removeElement(element)
		return key, true
	}
	return "", false
}
func (this *LRU) GetKeys() []string {
	keys := make([]string, len(this.Keys))
	i := 0
	for ent := this.List.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(string)
		i++
	}
	return keys
}
func (this *LRU) Len() int {
	return this.List.Len()
}
func (this *LRU) removeOldest() {
	element := this.List.Back()
	if element != nil {
		this.removeElement(element)
	}
}
func (this *LRU) removeElement(e *list.Element) {
	this.List.Remove(e)
	delete(this.Keys, e.Value.(string))
}

func (this *LRU) checkList() error {
	return nil
}
