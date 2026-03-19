package db

import (
	"sync"

	"goredis/internal/eviction"
)

// DB 是 Redis 键值存储的核心引擎。
//
// 职责：
//   1. 管理键空间（map[string]*Object）
//   2. 跟踪内存使用（usedBytes）
//   3. 触发 LRU 淘汰（当 usedBytes > maxMemoryBytes 时）
//   4. 提供 SET、GET、DEL、EXPIRE、TTL、KEYS 等操作
//   5. 通过 RWMutex 保证并发安全
//
// 内存模型：
//   - data map 存储实际的键值对
//   - lru 链表记录访问顺序
//   - usedBytes 累积所有 key + value 的大小（每项 +64 字节开销）
//
// 淘汰流程：
//   1. SET 一个新值，usedBytes 增加
//   2. 检测 usedBytes > maxMemoryBytes
//   3. 从 LRU 尾部循环 Evict，删除数据，减少 usedBytes
//   4. 直到 usedBytes <= maxMemoryBytes
type DB struct {
	// mu 读写锁，保护 data、lru、usedBytes 的并发访问
	// - 写操作（SET、DEL、EXPIRE）使用 Lock()
	// - 读操作（GET、TTL、KEYS）使用 RLock() 以支持并发读
	mu sync.RWMutex

	// data 是键空间：key -> Object
	// Object 包含值、过期时间等元数据
	data map[string]*Object

	// lru 访问顺序跟踪器（双向链表 + map）
	// - Touch(key) 时，key 移到"最近使用"位置（链表头）
	// - Evict() 时，从"最久未使用"位置删除（链表尾）
	lru *eviction.LRU

	// maxMemoryBytes 内存上限（单位：字节）
	// 0 = 无上限（不触发淘汰）
	// > 0 = 触发淘汰的阈值
	maxMemoryBytes int64

	// usedBytes 当前已使用的内存量
	// 累积 approxSize(key, value) = len(key) + len(value) + 64
	// 用来判断是否需要淘汰：if usedBytes > maxMemoryBytes { evict() }
	usedBytes int64
}

// New 创建一个无内存限制的 DB。
func New() *DB {
	return NewWithOptions(0)
}

// NewWithOptions 创建一个指定内存上限的 DB。
// 当 maxMemoryBytes > 0 时，超过上限会自动触发 LRU 淘汰。
func NewWithOptions(maxMemoryBytes int64) *DB {
	return &DB{
		data:           make(map[string]*Object),
		lru:            eviction.NewLRU(),
		maxMemoryBytes: maxMemoryBytes,
	}
}
