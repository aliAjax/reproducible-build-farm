package cache

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"sync"
	"time"
)

const defaultTTL = 24 * time.Hour

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
	v, ok := m.entries[k]
	m.mu.RUnlock()
	if !ok || v.Negative {
		return v, false
	}
	if expired(v) {
		m.mu.Lock()
		if cur, ok := m.entries[k]; ok && expired(cur) {
			delete(m.entries, k)
		}
		m.mu.Unlock()
		return v, false
	}
	return v, true
}
func (m *Memory) Put(_ context.Context, v domain.CacheEntry) error {
	if v.ExpiresAt.IsZero() {
		v.ExpiresAt = time.Now().Add(defaultTTL)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, e := range m.entries {
		if expired(e) {
			delete(m.entries, k)
		}
	}
	if len(m.entries) >= m.max {
		var victim digest.Digest
		var oldest time.Time
		for k, e := range m.entries {
			if victim == "" || e.CreatedAt.Before(oldest) {
				victim = k
				oldest = e.CreatedAt
			}
		}
		delete(m.entries, victim)
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

func expired(e domain.CacheEntry) bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}
