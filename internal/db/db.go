package db

import (
	"sync"

	"goredis/internal/eviction"
)

// shard 是原来 DB 的化身，现在它只是 16 个分片中的一个
type shard struct {
	mu             sync.RWMutex
	data           map[string]*Object
	lru            *eviction.LRU
	maxMemoryBytes int64
	usedBytes      int64
}

// newShard 创建一个独立的分片
func newShard(maxMemoryBytes int64) *shard {
	return &shard{
		data:           make(map[string]*Object),
		lru:            eviction.NewLRU(),
		maxMemoryBytes: maxMemoryBytes,
	}
}

// ================== 下面是新增的真正的 DB (路由层) ==================

const shardCount = 16

// DB 是暴露给外部的真数据库引擎，它内部管理着 16 个分片
type DB struct {
	shards []*shard
}

// fnv32 哈希算法：极其轻量、运算极快，非常适合做 key 的分片路由
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= 16777619
		hash ^= uint32(key[i])
	}
	return hash
}

// getShard 核心路由方法：根据 key 算出哈希，决定去哪个分片
func (d *DB) getShard(key string) *shard {
	idx := fnv32(key) % shardCount
	return d.shards[idx]
}

// NewWithOptions 创建包含 16 个分段锁的 DB
func NewWithOptions(maxMemoryBytes int64) *DB {
	d := &DB{
		shards: make([]*shard, shardCount),
	}
	// 将总内存平分给 16 个分片
	shardMaxMemory := int64(0)
	if maxMemoryBytes > 0 {
		shardMaxMemory = maxMemoryBytes / shardCount
	}
	for i := 0; i < shardCount; i++ {
		d.shards[i] = newShard(shardMaxMemory)
	}
	return d
}

// New 创建一个无内存限制的 DB
func New() *DB {
	return NewWithOptions(0)
}