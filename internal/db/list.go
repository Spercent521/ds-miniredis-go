package db

import "time"

// LPush 将一个或多个值插入到列表的头部。
// 返回插入后列表的长度。
func (d *DB) LPush(key string, values ...string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 查找或创建 Object
	obj, ok := d.data[key]
	if !ok {
		// Key 不存在，初始化一个新的 List 对象
		obj = &Object{
			Type:       ListType,
			List:       make([]string, 0),
			ExpireAtMs: 0,
			LastAccess: time.Now().UnixMilli(),
		}
		d.data[key] = obj
		// 新增 Key 的基础开销
		d.usedBytes += entryOverheadBytes 
	} else if obj.Type != ListType {
		// 类型冲突：比如针对一个 String 执行 LPUSH，直接返回 0 (严谨点应该返回 error)
		return 0
	}

	// 2. 将数据插入头部并计算内存
	var addedSize int64
	// Redis 的 LPUSH key v1 v2 v3，最终在左侧的顺序是 v3 v2 v1
	for _, v := range values {
		// Go 切片头部插入的经典写法：
		// 将新元素 v 放在最前面，把原来的 obj.List 展开跟在后面
		obj.List = append([]string{v}, obj.List...)
		addedSize += int64(len(v))
	}

	// 3. 更新系统状态
	d.usedBytes += addedSize
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key) // 标记为最近使用

	// 4. 触发内存淘汰 (复用之前 SetString 里的逻辑)
	d.checkAndEvict() 

	return len(obj.List)
}

// 辅助函数：抽取出来的内存淘汰逻辑，方便复用
func (d *DB) checkAndEvict() {
	if d.maxMemoryBytes > 0 {
		for d.usedBytes > d.maxMemoryBytes {
			evictKey, ok := d.lru.Evict()
			if !ok {
				break
			}
			if obj, exists := d.data[evictKey]; exists {
				// 扣除淘汰对象的内存（根据类型判断扣除量）
				if obj.Type == StringType {
					d.usedBytes -= approxSize(evictKey, obj.Str)
				} else if obj.Type == ListType {
					size := entryOverheadBytes + int64(len(evictKey))
					for _, v := range obj.List {
						size += int64(len(v))
					}
					d.usedBytes -= size
				}
				delete(d.data, evictKey)
			}
		}
	}
}

// LPop 移除并返回列表的第一个元素。
// 返回值：(元素值, 是否成功弹出)
func (d *DB) LPop(key string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 查找 Key
	obj, ok := d.data[key]
	if !ok || obj.Type != ListType || len(obj.List) == 0 {
		return "", false
	}

	// 2. 惰性过期检查
	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		// 已过期，清理内存并删除
		d.usedBytes -= d.calculateListSize(key, obj)
		d.lru.Remove(key)
		delete(d.data, key)
		return "", false
	}

	// 3. 弹出左侧（头部）第一个元素
	val := obj.List[0]
	// 切片操作：保留索引 1 到末尾的所有元素
	obj.List = obj.List[1:]

	// 4. 更新状态
	d.usedBytes -= int64(len(val)) // 扣除弹出的元素内存
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	// 5. 【关键机制】如果列表空了，自动删除该 Key
	if len(obj.List) == 0 {
		d.usedBytes -= entryOverheadBytes + int64(len(key)) // 扣除 Key 本身的开销
		d.lru.Remove(key)
		delete(d.data, key)
	}

	return val, true
}

// 辅助函数：计算整个 List 占用的内存
func (d *DB) calculateListSize(key string, obj *Object) int64 {
	size := entryOverheadBytes + int64(len(key))
	for _, v := range obj.List {
		size += int64(len(v))
	}
	return size
}