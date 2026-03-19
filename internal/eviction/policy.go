// Package eviction 实现内存淘汰策略。
//
// Redis 在内存超过上限时需要删除一些数据。
// 本包提供 LRU（最近最少使用）策略的实现。
//
// 【推荐开始阅读的文件】 👈 从此文件开始
// 1. 本文件 (policy.go) - 定义淘汰策略接口
// 2. lru.go - 双向链表实现的 LRU
// 3. ../object.go - 键值对数据结构
// 4. ../db.go - 核心引擎（集成 LRU）
// 5. ../string.go - SET/GET 等字符串操作
// 6. ../generic.go - EXPIRE、TTL 等通用操作
package eviction

// Policy 定义内存淘汰策略的通用接口。
// 任何淘汰算法都必须实现这三个方法。
type Policy interface {
	// Touch 更新 key 的访问位置。
	// 在 SET、GET 等操作后调用，标记这个 key 最近被使用。
	Touch(key string)

	// Evict 淘汰一个最不活跃的 key。
	// 返回被淘汰 key 的名称，以及是否成功（若无可淘汰的 key，ok=false）。
	Evict() (string, bool)

	// Remove 从跟踪中移除一个 key。
	// 在 DELETE、EXPIRE 后调用，清理其访问记录。
	Remove(key string)
}
