package expire

import "time"

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) UnixMilliAfterTTL(ttl time.Duration) int64 {
	return time.Now().Add(ttl).UnixMilli()
}
