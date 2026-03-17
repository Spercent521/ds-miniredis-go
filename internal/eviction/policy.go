package eviction

type Policy interface {
	Touch(key string)
	Evict() (string, bool)
	Remove(key string)
}
