package db

import (
	"sort"
	"strconv"
	"time"
)

// ZSetMember 用于在排序时临时组合 member 和 score
type ZSetMember struct {
	Member string
	Score  float64
}

// ZAdd 将一个或多个 member 元素及其 score 值加入到有序集 key 当中。
func (d *shard) ZAdd(key string, pairs ...ZSetMember) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 查找或创建 Object
	obj, ok := d.data[key]
	if !ok {
		obj = &Object{
			Type:       ZSetType,
			ZSet:       make(map[string]float64),
			ExpireAtMs: 0,
			LastAccess: time.Now().UnixMilli(),
		}
		d.data[key] = obj
		d.usedBytes += entryOverheadBytes
	} else if obj.Type != ZSetType {
		return 0
	}

	// 2. 遍历写入数据
	added := 0
	for _, p := range pairs {
		if _, exists := obj.ZSet[p.Member]; !exists {
			// 新增元素，增加内存估算
			added++
			d.usedBytes += int64(len(p.Member)) + 64
		}
		// 无论是否存在，都更新为最新的 score
		obj.ZSet[p.Member] = p.Score
	}

	// 3. 更新状态
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)

	return added
}

// ZRange 返回有序集中，指定区间内的成员。
func (d *shard) ZRange(key string, start, stop int, withScores bool) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	obj, ok := d.data[key]
	if !ok || obj.Type != ZSetType {
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

	// --- 核心排序逻辑 ---
	// 1. 提取到切片中
	members := make([]ZSetMember, 0, len(obj.ZSet))
	for m, s := range obj.ZSet {
		members = append(members, ZSetMember{Member: m, Score: s})
	}

	// 2. 根据 Score 进行升序排序（如果 Score 相同，则按字典序排）
	sort.Slice(members, func(i, j int) bool {
		if members[i].Score == members[j].Score {
			return members[i].Member < members[j].Member
		}
		return members[i].Score < members[j].Score
	})

	// 3. 处理 Redis 经典的负数索引（-1 代表最后一个元素）
	size := len(members)
	if start < 0 {
		start = size + start
	}
	if stop < 0 {
		stop = size + stop
	}
	if start < 0 {
		start = 0
	}
	if start > stop || start >= size {
		return nil
	}
	if stop >= size {
		stop = size - 1
	}

	// 4. 组装返回结果
	var result []string
	for i := start; i <= stop; i++ {
		result = append(result, members[i].Member)
		if withScores {
			// 将 float64 转换为字符串返回
			scoreStr := strconv.FormatFloat(members[i].Score, 'f', -1, 64)
			result = append(result, scoreStr)
		}
	}

	return result
}