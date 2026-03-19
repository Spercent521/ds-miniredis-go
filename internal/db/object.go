// Package db 实现键值存储引擎的核心数据结构和操作。
//
// 这个包集成了：
//   - 内存中的键空间管理
//   - TTL（过期时间）实现
//   - LRU 内存淘汰
//   - 字符串值存储
//   - 通用操作（KEYS、EXPIRE、TTL 等）
//
// 【核心流程】
// 1. Object - 键值对的统一数据结构（可扩展为 List、Hash 等）
// 2. DB - 键空间、LRU 链、内存计数
// 3. SetString - 写值，同时触发 LRU Touch 和内存淘汰
// 4. GetString - 读值，同时检查过期和刷新 LRU
package db

// ObjectType 表示存储对象的类型。
// Redis 支持多种数据类型，当前项目只实现了 String。
type ObjectType uint8

const (
	// StringType 表示字符串值
	StringType ObjectType = iota
	// ListType 表示列表值
	ListType
	// 哈希字典
	HashType 
	// 无序集合
	SetType  
	// 有序集合
	ZSetType 
)

// Object 是 Redis 存储的统一数据结构。
// 每个 key 对应一个 Object，包含值、过期时间等元数据。
type Object struct {
	// Type 对象的类型（当前项目只支持 StringType）
	Type ObjectType

	// Str 字符串值的内容
	Str string

	// List 列表值的内容（当 Type == ListType 时使用）
	// 使用 Go 原生切片作为基础实现
	List []string

	// Hash 使用 Go 的 map 实现，存储 field -> value
	Hash map[string]string 
	
	// Set 使用 Go 的 map[string]struct{} 实现（struct{} 不占内存，完美充当集合去重）
	Set  map[string]struct{} 
	
	// ZSet 暂时用 map 存储 member -> score，后续实现排序逻辑
	ZSet map[string]float64

	// ExpireAtMs 过期时间（Unix 毫秒戳）
	// 0 = 永不过期（无 TTL）
	// > 0 = 在此时刻过期
	// 过期检查采用"惰性删除"：只在访问时检查，不主动清理
	ExpireAtMs int64

	// LastAccess 最后一次访问时间（Unix 毫秒戳）
	// 用于 LRU 计算，但当前并未在淘汰时实际使用
	// （淘汰用的是链表顺序，而非此时间戳）
	LastAccess int64
}
