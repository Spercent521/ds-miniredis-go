package db

import "time"

// SAdd 将一个或多个 member 元素加入到集合 key 当中。
// 已经存在于集合的 member 元素将被忽略。
// 返回值：被添加到集合中的新元素的数量，不包括被忽略的元素。
func (d *shard) SAdd(key string, members ...string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 查找或创建 Object
	obj, ok := d.data[key]
	if !ok {
		obj = &Object{
			Type:       SetType,
			Set:        make(map[string]struct{}),
			ExpireAtMs: 0,
			LastAccess: time.Now().UnixMilli(),
		}
		d.data[key] = obj
		d.usedBytes += entryOverheadBytes
	} else if obj.Type != SetType {
		// 类型冲突
		return 0
	}

	// 2. 遍历 members 并去重插入
	added := 0
	for _, member := range members {
		if _, exists := obj.Set[member]; !exists {
			// 插入空结构体作为值
			obj.Set[member] = struct{}{}
			// 估算内存：字符串长度 + map 节点大概开销(按64算)
			d.usedBytes += int64(len(member)) + 64 
			added++
		}
	}

	// 3. 更新 LRU 和访问时间
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	// 4. 触发内存淘汰 (如果配置了 maxMemoryBytes)
	if d.maxMemoryBytes > 0 {
		// d.checkAndEvict() 
	}

	return added
}

// SMembers 返回集合 key 中的所有成员。
func (d *shard) SMembers(key string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	obj, ok := d.data[key]
	if !ok || obj.Type != SetType {
		return nil
	}

	// 惰性过期检查
	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		d.lru.Remove(key)
		delete(d.data, key)
		return nil
	}

	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	// 提取 map 的所有 key 形成切片
	result := make([]string, 0, len(obj.Set))
	for member := range obj.Set {
		result = append(result, member)
	}
	return result
}

// SIsMember 判断 member 元素是否是集合 key 的成员。
// 返回值：如果是返回 1，否则返回 0。
func (d *shard) SIsMember(key, member string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	obj, ok := d.data[key]
	if !ok || obj.Type != SetType {
		return 0
	}

	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		d.lru.Remove(key)
		delete(d.data, key)
		return 0
	}

	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	if _, exists := obj.Set[member]; exists {
		return 1
	}
	return 0
}