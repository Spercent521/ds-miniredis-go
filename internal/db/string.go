package db

import "time"

// entryOverheadBytes 是每个 key-value 项的额外内存开销（单位：字节）
// 考虑 Go 的对象头、内部指针等，估算为 64 字节
const entryOverheadBytes = 64

// approxSize 估算一个 key-value 对占用的内存大小。
// 公式：len(key) + len(value) + 固定开销(64字节)
func approxSize(key, value string) int64 {
	return int64(len(key)+len(value)) + entryOverheadBytes
}

// SetString 写入一个字符串值，并触发 LRU 与内存淘汰。
//
// 参数：
//   - key: 键名
//   - value: 值内容
//   - expireAtMs: 过期时间（Unix 毫秒戳），0 表示永不过期
//
// 执行步骤：
//   1. 若 key 已存在，先扣除旧值的内存占用，从 LRU 移除
//   2. 写入新对象到 data map
//   3. 增加 usedBytes 计数
//   4. 调用 lru.Touch() 标记为"最近使用"
//   5. 若 usedBytes > maxMemoryBytes，循环淘汰最久未使用的 key
func (d *shard) SetString(key, value string, expireAtMs int64) {
	// 加写锁保护并发修改
	d.mu.Lock()
	defer d.mu.Unlock()

	// === 步骤 1：处理旧值（若 key 已存在，则覆盖） ===
	if old, ok := d.data[key]; ok {
		// key 已存在，需要先扣除旧值占用的内存
		d.usedBytes -= approxSize(key, old.Str)
		// 从 LRU 链表中移除旧值
		d.lru.Remove(key)
	}

	// === 步骤 2：写入新对象 ===
	d.data[key] = &Object{
		Type:       StringType,
		Str:        value,
		ExpireAtMs: expireAtMs, // 0 = 无 TTL，> 0 = 绝对时刻
		LastAccess: time.Now().UnixMilli(),
	}

	// === 步骤 3：更新内存计数 ===
	d.usedBytes += approxSize(key, value)

	// === 步骤 4：LRU Touch（标记为"最近使用"） ===
	d.lru.Touch(key)

	// === 步骤 5：内存淘汰（若超过上限） ===
	if d.maxMemoryBytes > 0 {
		// 当 usedBytes > maxMemoryBytes 时，循环淘汰
		for d.usedBytes > d.maxMemoryBytes {
			// 从 LRU 尾部取出最久未使用的 key
			evictKey, ok := d.lru.Evict()
			if !ok {
				// LRU 链表空了，无法继续淘汰，退出
				break
			}
			// 从 data map 中删除该 key
			if obj, exists := d.data[evictKey]; exists {
				// 扣除淘汰 key 占用的内存
				d.usedBytes -= approxSize(evictKey, obj.Str)
				delete(d.data, evictKey)
			}
		}
	}
}

// GetString 读取字符串值，同时进行"惰性"过期检查和 LRU 刷新。n//
// 返回值：
//   - (value, true) - key 存在且未过期
//   - ("", false) - key 不存在或已过期
//
// 执行步骤：
//   1. 查找 key 是否存在
//   2. 检查是否已过期（ExpireAtMs > 0 && current_time > ExpireAtMs）
//   3. 若过期，立即删除（惰性删除，不主动清理）
//   4. 若未过期，更新 LastAccess，调用 lru.Touch() 刷新位置
func (d *shard) GetString(key string) (string, bool) {
	// 加写锁保护并发修改（检查过期并删除需要加锁）
	d.mu.Lock()
	defer d.mu.Unlock()

	// === 步骤 1：查找 key ===
	obj, ok := d.data[key]
	if !ok || obj.Type != StringType {
		// key 不存在或类型不是 String，返回"不存在"
		return "", false
	}

	// === 步骤 2：检查是否过期（惰性检查） ===
	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		// 已过期，立即删除
		d.usedBytes -= approxSize(key, obj.Str)
		d.lru.Remove(key)
		delete(d.data, key)
		return "", false
	}

	// === 步骤 3：刷新 LRU 和访问时间 ===
	// 表示这个 key "刚被访问"，更新 LRU 链表顺序
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key) // 将 key 移到链表头（最近使用）

	// === 返回值 ===
	return obj.Str, true
}

// Del 删除一个或多个 key，返回成功删除的个数。
func (d *shard) Del(keys ...string) int {
	// 加写锁保护并发修改
	d.mu.Lock()
	defer d.mu.Unlock()
	deleted := 0
	for _, key := range keys {
		if obj, ok := d.data[key]; ok {
			// key 存在，执行删除
			d.usedBytes -= approxSize(key, obj.Str) // 扣除内存
			d.lru.Remove(key)                      // 从 LRU 链表移除
			delete(d.data, key)                   // 从 data map 删除
			deleted++
		}
	}
	return deleted
}
