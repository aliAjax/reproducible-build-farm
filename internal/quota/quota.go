package quota

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"sync"
)

type Manager struct {
	mu    sync.Mutex
	used  domain.ResourceBudget
	limit domain.ResourceBudget
}

func New(limit domain.ResourceBudget) *Manager { return &Manager{limit: limit} }
func (m *Manager) Reserve(_ context.Context, r domain.ResourceBudget) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.limit.CPU > 0 && m.used.CPU+r.CPU > m.limit.CPU {
		return fmt.Errorf("cpu quota exceeded")
	}
	if m.limit.MemoryMB > 0 && m.used.MemoryMB+r.MemoryMB > m.limit.MemoryMB {
		return fmt.Errorf("memory quota exceeded")
	}
	m.used.CPU += r.CPU
	m.used.MemoryMB += r.MemoryMB
	return nil
}
func (m *Manager) Release(_ context.Context, r domain.ResourceBudget) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.used.CPU -= r.CPU
	m.used.MemoryMB -= r.MemoryMB
	if m.used.CPU < 0 {
		m.used.CPU = 0
	}
	if m.used.MemoryMB < 0 {
		m.used.MemoryMB = 0
	}
}
