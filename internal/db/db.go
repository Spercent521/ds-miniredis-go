package db

import (
	"sync"

	"goredis/internal/eviction"
)

type DB struct {
	mu             sync.RWMutex
	data           map[string]*Object
	lru            *eviction.LRU
	maxMemoryBytes int64
	usedBytes      int64
}

func New() *DB {
	return NewWithOptions(0)
}

func NewWithOptions(maxMemoryBytes int64) *DB {
	return &DB{
		data:           make(map[string]*Object),
		lru:            eviction.NewLRU(),
		maxMemoryBytes: maxMemoryBytes,
	}
}
