package node

import (
	"sync"
)

type DB struct {
	mutex sync.RWMutex
	data  map[string]string
}

func newDB() *DB {
	return &DB{
		data: make(map[string]string),
	}
}
func (db *DB) Data()map[string]string {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.data
}
func (db *DB) SetData(data map[string]string) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	db.data=data
}

func (db *DB) Set(key string, value string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.data[key] = value
}
func (db *DB) Get(key string) string {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.data[key]
}
