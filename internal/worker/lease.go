package worker

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu    sync.Mutex
	store repository.Store
	ttl   time.Duration
	seq   int64
}

func NewManager(s repository.Store, ttl time.Duration) *Manager {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Manager{store: s, ttl: ttl}
}
func (m *Manager) Acquire(ctx context.Context, execID string, budget domain.ResourceBudget) (domain.Lease, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.store.ListWorkers(ctx) {
		if !w.Busy && w.Capacity.CPU >= budget.CPU && w.Capacity.MemoryMB >= budget.MemoryMB {
			m.seq++
			id := fmt.Sprintf("lease-%d", m.seq)
			w.Busy = true
			w.LeaseID = id
			_ = m.store.SaveWorker(ctx, w)
			return domain.Lease{ID: id, ExecutionID: execID, WorkerID: w.ID, Version: m.seq, ExpiresAt: time.Now().Add(m.ttl)}, nil
		}
	}
	return domain.Lease{}, fmt.Errorf("no worker capacity")
}
func (m *Manager) Release(ctx context.Context, l domain.Lease) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, err := m.store.GetWorker(ctx, l.WorkerID)
	if err != nil {
		return err
	}
	if w.LeaseID != l.ID {
		return domain.ErrLeaseLost
	}
	w.Busy = false
	w.LeaseID = ""
	return m.store.SaveWorker(ctx, w)
}
