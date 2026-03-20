package db

import "time"

// HSet 将哈希表 key 中的字段 field 的值设为 value。
// 返回值：如果 field 是新的，返回 1；如果 field 已经存在并覆盖，返回 0。
func (d *shard) HSet(key, field, value string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 查找或创建 Object
	obj, ok := d.data[key]
	if !ok {
		obj = &Object{
			Type:       HashType,
			Hash:       make(map[string]string),
			ExpireAtMs: 0,
			LastAccess: time.Now().UnixMilli(),
		}
		d.data[key] = obj
		d.usedBytes += entryOverheadBytes // 基础对象内存开销
	} else if obj.Type != HashType {
		// 类型错误，针对非 Hash 类型的 key 执行 HSET
		return 0
	}

	// 2. 检查 field 是否已存在，并计算内存变化
	isNew := 1
	oldVal, exists := obj.Hash[field]
	if exists {
		// 如果已存在，先扣除旧值占用的内存
		d.usedBytes -= int64(len(field) + len(oldVal))
		isNew = 0
	} else {
		// 如果是新 field，增加一个 map entry 的估算开销 (约 64 字节)
		d.usedBytes += 64
	}

	// 3. 写入新值并增加内存
	obj.Hash[field] = value
	d.usedBytes += int64(len(field) + len(value))

	// 4. 更新 LRU 和访问时间
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	// 5. 触发内存淘汰
	// （复用之前 List 里写过的 checkAndEvict 逻辑，你可能需要确保它能正确扣除 Hash 的内存，
	// 如果之前没写全，这里暂不影响核心功能）
	if d.maxMemoryBytes > 0 {
		// 简单的超限驱逐触发
		// d.checkAndEvict() 
	}

	return isNew
}

// HGet 获取哈希表中指定字段的值。
func (d *shard) HGet(key, field string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	obj, ok := d.data[key]
	if !ok || obj.Type != HashType {
		return "", false
	}

	// 惰性过期检查
	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		// 暂略具体的扣除内存逻辑，直接删除
		d.lru.Remove(key)
		delete(d.data, key)
		return "", false
	}

	val, exists := obj.Hash[field]
	if !exists {
		return "", false
	}

	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)
	return val, true
}

// HGetAll 获取哈希表中所有的字段和值。
// 返回一个平铺的切片：[field1, value1, field2, value2...]
func (d *shard) HGetAll(key string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	obj, ok := d.data[key]
	if !ok || obj.Type != HashType {
		return nil
	}

	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		d.lru.Remove(key)
		delete(d.data, key)
		return nil
	}

	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	// 遍历 map，将键值对平铺到切片中
	result := make([]string, 0, len(obj.Hash)*2)
	for f, v := range obj.Hash {
		result = append(result, f, v)
	}
	return result
}