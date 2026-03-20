package db

import (
	"path"
	"sort"
	"time"

	"goredis/internal/eviction"
)

// Expire 为一个已存在的 key 设置过期时间（绝对时刻，Unix 毫秒）。
//
// 参数：
//   - key: 键名
//   - expireAtMs: 过期时刻（Unix 毫秒戳）
//
// 返回：
//   - true 若 key 存在且设置成功
//   - false 若 key 不存在
func (d *shard) Expire(key string, expireAtMs int64) bool {
	// 加写锁（修改 ExpireAtMs）
	d.mu.Lock()
	defer d.mu.Unlock()
	obj, ok := d.data[key]
	if !ok {
		return false
	}
	// 直接修改对象的过期时间
	obj.ExpireAtMs = expireAtMs
	return true
}

// TTLMs 返回一个 key 剩余的生存时间（毫秒）。
//
// 返回值说明：
//   - > 0: 还有多少毫秒才过期
//   - -1: key 存在但没有设置过期时间（永久有效）
//   - -2: key 不存在或已过期
//
// 这个方法不会删除过期 key，只是计算剩余时间。
// 删除过期 key 发生在 GetString（惰性删除）中。
func (d *shard) TTLMs(key string) int64 {
	// 加读锁（只读操作）
	d.mu.RLock()
	defer d.mu.RUnlock()

	// === 检查 key 是否存在 ===
	obj, ok := d.data[key]
	if !ok {
		return -2 // key 不存在
	}

	// === 检查是否设置过过期时间 ===
	if obj.ExpireAtMs == 0 {
		return -1 // 无过期时间（永久有效）
	}

	// === 计算剩余时间 ===
	rem := obj.ExpireAtMs - time.Now().UnixMilli()
	if rem <= 0 {
		return -2 // 已过期
	}
	return rem // 剩余毫秒数
}

// Exists 返回给定的 keys 中有多少个存在且未过期。
func (d *shard) Exists(keys ...string) int {
	// 加读锁（只读操作）
	d.mu.RLock()
	defer d.mu.RUnlock()
	now := time.Now().UnixMilli()
	n := 0
	for _, key := range keys {
		obj, ok := d.data[key]
		if !ok {
			continue // key 不存在，跳过
		}
		// 检查是否已过期（惰性检查，不删除）
		if obj.ExpireAtMs > 0 && now > obj.ExpireAtMs {
			continue // 已过期，不计数
		}
		n++ // 存在且未过期，计数 +1
	}
	return n
}

// Keys 返回所有匹配给定 glob 模式的 key（不含过期 key）。
// 结果按字母排序，确保输出的一致性。
//
// pattern 支持通配符：
//   - * 匹配任意字符序列
//   - ? 匹配单个字符
//   - [abc] 匹配字符集合
//
// 示例：
//   - "user:*" 匹配 user:1、user:2 等
//   - "key?" 匹配 key1、keyA 等
func (d *shard) Keys(pattern string) []string {
	// 加读锁（只读操作）
	d.mu.RLock()
	defer d.mu.RUnlock()
	now := time.Now().UnixMilli()
	var result []string
	for key, obj := range d.data {
		// 跳过已过期的 key（惰性检查）
		if obj.ExpireAtMs > 0 && now > obj.ExpireAtMs {
			continue
		}
		// 检查 key 是否匹配 pattern
		matched, err := path.Match(pattern, key)
		if err == nil && matched {
			result = append(result, key)
		}
	}
	// 排序确保输出顺序一致（重要的是确定性，便于测试）
	sort.Strings(result)
	return result
}

// DBSize 返回当前DB中key的数量（包含未被惰性删除的过期key）。
// 注意：这个数字可能包含已过期但尚未被访问删除的key。
// 如果需要精确的非过期key数量，需要另外计数。
func (d *shard) DBSize() int {
	// 加读锁（只读操作）
	d.mu.RLock()
	defer d.mu.RUnlock()
	// 直接返回 data map 的长度，O(1) 操作
	return len(d.data)
}

// FlushDB 清空所有数据并重置内存计数。
// 通常在 FLUSHDB 命令或 AOF 重置时调用。
//
// 执行步骤：
//   1. 清空 data map（所有 key-value 对）
//   2. 重建 LRU 链表（清除所有访问记录）
//   3. 重置 usedBytes 为 0（内存计数清零）
func (d *shard) FlushDB() {
	// 加写锁（修改所有数据）
	d.mu.Lock()
	defer d.mu.Unlock()
	// 创建新的空 map，丢弃旧 map（Go 会自动GC）
	d.data = make(map[string]*Object)
	// 重建 LRU 链表
	d.lru = eviction.NewLRU()
	// 内存计数清零
	d.usedBytes = 0
}
