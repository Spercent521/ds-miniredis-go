package db

import (
	"path"
	"sort"
	"time"

	"goredis/internal/eviction"
)

// Expire sets an absolute expiry timestamp (Unix ms) on an existing key.
// Returns false if the key does not exist.
func (d *DB) Expire(key string, expireAtMs int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	obj, ok := d.data[key]
	if !ok {
		return false
	}
	obj.ExpireAtMs = expireAtMs
	return true
}

// TTLMs returns the remaining TTL of key in milliseconds.
//
//	-1  → key exists and has no expiry
//	-2  → key does not exist or has already expired
func (d *DB) TTLMs(key string) int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	obj, ok := d.data[key]
	if !ok {
		return -2
	}
	if obj.ExpireAtMs == 0 {
		return -1
	}
	rem := obj.ExpireAtMs - time.Now().UnixMilli()
	if rem <= 0 {
		return -2
	}
	return rem
}

// Exists returns the count of provided keys that are present and not expired.
func (d *DB) Exists(keys ...string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	now := time.Now().UnixMilli()
	n := 0
	for _, key := range keys {
		obj, ok := d.data[key]
		if !ok {
			continue
		}
		if obj.ExpireAtMs > 0 && now > obj.ExpireAtMs {
			continue
		}
		n++
	}
	return n
}

// Keys returns all non-expired keys matching the glob pattern (path.Match syntax).
// Results are sorted for deterministic output.
func (d *DB) Keys(pattern string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	now := time.Now().UnixMilli()
	var result []string
	for key, obj := range d.data {
		if obj.ExpireAtMs > 0 && now > obj.ExpireAtMs {
			continue
		}
		matched, err := path.Match(pattern, key)
		if err == nil && matched {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}

// DBSize returns the number of keys currently in the store (may include lazily-expired keys).
func (d *DB) DBSize() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.data)
}

// FlushDB removes all keys and resets memory counters.
func (d *DB) FlushDB() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data = make(map[string]*Object)
	d.lru = eviction.NewLRU()
	d.usedBytes = 0
}
