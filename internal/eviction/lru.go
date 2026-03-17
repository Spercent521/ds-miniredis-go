package eviction

import "container/list"

type LRU struct {
	list  *list.List
	nodes map[string]*list.Element
}

type entry struct {
	key string
}

func NewLRU() *LRU {
	return &LRU{list: list.New(), nodes: make(map[string]*list.Element)}
}

func (l *LRU) Touch(key string) {
	if n, ok := l.nodes[key]; ok {
		l.list.MoveToFront(n)
		return
	}
	n := l.list.PushFront(entry{key: key})
	l.nodes[key] = n
}

func (l *LRU) Evict() (string, bool) {
	back := l.list.Back()
	if back == nil {
		return "", false
	}
	e := back.Value.(entry)
	delete(l.nodes, e.key)
	l.list.Remove(back)
	return e.key, true
}

func (l *LRU) Remove(key string) {
	if n, ok := l.nodes[key]; ok {
		delete(l.nodes, key)
		l.list.Remove(n)
	}
}
