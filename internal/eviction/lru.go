package eviction

import "container/list"

// LRU 实现了最近最少使用（Least Recently Used）淘汰策略。
//
// 设计原理：
//   - 维护一个双向链表，Front（头部）是最近使用的，Back（尾部）是最久未使用的
//   - 每当 Touch 某个 key，将其移到 Front
//   - 淘汰时，选尾部的 key 删除
//   - 通过 map 快速定位 key 在链表中哪个位置，O(1) 操作
//
// 内存消耗：每个 key 占用一个链表节点 + map 条目，开销很小。
type LRU struct {
	// list: 双向链表，维护访问的时间顺序
	// - Front (头) = 最近使用
	// - Back (尾) = 最久未使用 ← 淘汰时删这个
	list *list.List

	// nodes: key -> 链表节点 的映射，用来快速定位和移动
	// 不用遍历链表，直接 O(1) 找到节点位置
	nodes map[string]*list.Element
}

// entry 是链表中的一个节点，仅含 key（value 在 DB 中存储）。
type entry struct {
	key string
}

// NewLRU 创建一个新的 LRU 跟踪器。
func NewLRU() *LRU {
	return &LRU{list: list.New(), nodes: make(map[string]*list.Element)}
}

// Touch 标记 key 为"刚刚被访问"。
// 如果 key 已存在，将其移到链表头（最新位置）。
// 如果是新 key，插入列表头。
func (l *LRU) Touch(key string) {
	if n, ok := l.nodes[key]; ok {
		// key 已在链表中，将其移到头部（最近使用）
		l.list.MoveToFront(n)
		return
	}
	// key 是新的，加入链表头
	n := l.list.PushFront(entry{key: key})
	l.nodes[key] = n
}

// Evict 淘汰最久未使用的 key。
// 从链表尾部删除一个节点，返回被淘汰的 key 名称。
// 如果链表为空（没有 key 需要淘汰），返回 ("" , false)。
func (l *LRU) Evict() (string, bool) {
	// 取链表尾部（最久未使用）的节点
	back := l.list.Back()
	if back == nil {
		// 链表为空，无可淘汰
		return "", false
	}
	// 提取节点中的 key
	e := back.Value.(entry)
	// 从 map 和链表中同时删除
	delete(l.nodes, e.key)
	l.list.Remove(back)
	return e.key, true
}

// Remove 从 LRU 追踪中删除一个 key。
// 在删除数据或过期时调用，防止 LRU 链表与 DB 数据不同步。
func (l *LRU) Remove(key string) {
	if n, ok := l.nodes[key]; ok {
		// 从 map 和链表中同时删除
		delete(l.nodes, key)
		l.list.Remove(n)
	}
}
