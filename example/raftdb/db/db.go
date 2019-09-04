package db

import (
	"sync"
)

type DB struct {
	data  map[string]string
	mutex sync.RWMutex
}

func New() *DB {
	return &DB{
		data: make(map[string]string),
	}
}

func (db *DB) Get(key string) string {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.data[key]
}

func (db *DB) Put(key string, value string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.data[key] = value
}
