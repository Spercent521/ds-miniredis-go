package db

import "time"

const entryOverheadBytes = 64

func approxSize(key, value string) int64 {
	return int64(len(key)+len(value)) + entryOverheadBytes
}

// SetString writes a string value. expireAtMs == 0 means no expiry.
// After writing it touches LRU and evicts least-recently-used keys when over maxMemoryBytes.
func (d *DB) SetString(key, value string, expireAtMs int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Subtract old size if key already exists.
	if old, ok := d.data[key]; ok {
		d.usedBytes -= approxSize(key, old.Str)
		d.lru.Remove(key)
	}

	d.data[key] = &Object{
		Type:       StringType,
		Str:        value,
		ExpireAtMs: expireAtMs,
		LastAccess: time.Now().UnixMilli(),
	}
	d.usedBytes += approxSize(key, value)
	d.lru.Touch(key)

	// LRU eviction: free memory until under the limit.
	if d.maxMemoryBytes > 0 {
		for d.usedBytes > d.maxMemoryBytes {
			evictKey, ok := d.lru.Evict()
			if !ok {
				break
			}
			if obj, exists := d.data[evictKey]; exists {
				d.usedBytes -= approxSize(evictKey, obj.Str)
				delete(d.data, evictKey)
			}
		}
	}
}

// GetString returns the string value for key with lazy expiry and LRU touch.
func (d *DB) GetString(key string) (string, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	obj, ok := d.data[key]
	if !ok || obj.Type != StringType {
		return "", false
	}
	if obj.ExpireAtMs > 0 && time.Now().UnixMilli() > obj.ExpireAtMs {
		d.usedBytes -= approxSize(key, obj.Str)
		d.lru.Remove(key)
		delete(d.data, key)
		return "", false
	}
	obj.LastAccess = time.Now().UnixMilli()
	d.lru.Touch(key)
	return obj.Str, true
}

// Del removes the given keys and returns the count of actually deleted keys.
func (d *DB) Del(keys ...string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	deleted := 0
	for _, key := range keys {
		if obj, ok := d.data[key]; ok {
			d.usedBytes -= approxSize(key, obj.Str)
			d.lru.Remove(key)
			delete(d.data, key)
			deleted++
		}
	}
	return deleted
}
