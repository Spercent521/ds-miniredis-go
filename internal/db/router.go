package db

import "sort"

// ================= String 路由 =================
func (d *DB) SetString(key, value string, expireAtMs int64) {
	d.getShard(key).SetString(key, value, expireAtMs)
}

func (d *DB) GetString(key string) (string, bool) {
	return d.getShard(key).GetString(key)
}

// ================= List 路由 =================
func (d *DB) LPush(key string, values ...string) int {
	return d.getShard(key).LPush(key, values...)
}

func (d *DB) LPop(key string) (string, bool) {
	return d.getShard(key).LPop(key)
}

// ================= Hash 路由 =================
func (d *DB) HSet(key, field, value string) int {
	return d.getShard(key).HSet(key, field, value)
}

func (d *DB) HGet(key, field string) (string, bool) {
	return d.getShard(key).HGet(key, field)
}

func (d *DB) HGetAll(key string) []string {
	return d.getShard(key).HGetAll(key)
}

// ================= Set 路由 =================
func (d *DB) SAdd(key string, members ...string) int {
	return d.getShard(key).SAdd(key, members...)
}

func (d *DB) SMembers(key string) []string {
	return d.getShard(key).SMembers(key)
}

func (d *DB) SIsMember(key, member string) int {
	return d.getShard(key).SIsMember(key, member)
}

// ================= ZSet 路由 =================
func (d *DB) ZAdd(key string, pairs ...ZSetMember) int {
	return d.getShard(key).ZAdd(key, pairs...)
}

func (d *DB) ZRange(key string, start, stop int, withScores bool) []string {
	return d.getShard(key).ZRange(key, start, stop, withScores)
}

// ================= Generic 通用命令 (单键路由) =================
func (d *DB) Expire(key string, expireAtMs int64) bool {
	return d.getShard(key).Expire(key, expireAtMs)
}

func (d *DB) TTLMs(key string) int64 {
	return d.getShard(key).TTLMs(key)
}

// ================= Generic 通用命令 (多键或全局广播) =================

// Del 可能传入多个不同分片的 key，所以要拆解转发
func (d *DB) Del(keys ...string) int {
	deleted := 0
	for _, key := range keys {
		// 单独分发每一个 key 去对应分片删除
		deleted += d.getShard(key).Del(key)
	}
	return deleted
}

// Exists 也是多键，需要拆解转发
func (d *DB) Exists(keys ...string) int {
	count := 0
	for _, key := range keys {
		count += d.getShard(key).Exists(key)
	}
	return count
}

// Keys 需要去所有 16 个分片里找，然后汇总排序
func (d *DB) Keys(pattern string) []string {
	var result []string
	// 遍历所有分片
	for _, s := range d.shards {
		result = append(result, s.Keys(pattern)...)
	}
	// 合并后需要重新排序，确保全局输出一致性
	sort.Strings(result)
	return result
}

// DBSize 汇总所有分片的数量
func (d *DB) DBSize() int {
	total := 0
	for _, s := range d.shards {
		total += s.DBSize()
	}
	return total
}

// FlushDB 广播给所有分片执行清空
func (d *DB) FlushDB() {
	for _, s := range d.shards {
		s.FlushDB()
	}
}