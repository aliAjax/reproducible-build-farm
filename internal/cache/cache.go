package cache

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"sync"
	"time"
)

type Remote interface {
	Get(context.Context, digest.Digest) (domain.CacheEntry, bool)
	Put(context.Context, domain.CacheEntry) error
	Invalidate(context.Context, digest.Digest) error
}
type Memory struct {
	mu      sync.RWMutex
	entries map[digest.Digest]domain.CacheEntry
	max     int
}

func NewMemory(max int) *Memory {
	if max < 1 {
		max = 10000
	}
	return &Memory{entries: map[digest.Digest]domain.CacheEntry{}, max: max}
}
func (m *Memory) Get(_ context.Context, k digest.Digest) (domain.CacheEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.entries[k]
	if !ok || v.Negative {
		return v, false
	}
	if !v.ExpiresAt.IsZero() && time.Now().After(v.ExpiresAt) {
		return v, false
	}
	return v, true
}
func (m *Memory) Put(_ context.Context, v domain.CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) >= m.max {
		for k := range m.entries {
			delete(m.entries, k)
			break
		}
	}
	m.entries[v.ActionKey] = v
	return nil
}
func (m *Memory) Invalidate(_ context.Context, k digest.Digest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, k)
	return nil
}
