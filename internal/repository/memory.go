package repository

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"sync"
	"time"
)

type Store interface {
	CreateProject(context.Context, domain.Project) error
	GetProject(context.Context, string) (domain.Project, error)
	SaveDefinition(context.Context, domain.BuildDefinition) error
	GetDefinition(context.Context, string) (domain.BuildDefinition, error)
	SaveExecution(context.Context, domain.Execution) error
	GetExecution(context.Context, string) (domain.Execution, error)
	FindExecutionByIdempotency(context.Context, string) (domain.Execution, error)
	ListExecutions(context.Context) []domain.Execution
	SaveWorker(context.Context, domain.Worker) error
	GetWorker(context.Context, string) (domain.Worker, error)
	ListWorkers(context.Context) []domain.Worker
	PutCache(context.Context, domain.CacheEntry) error
	GetCache(context.Context, digest.Digest) (domain.CacheEntry, error)
	SaveAttestation(context.Context, domain.Attestation) error
	GetAttestation(context.Context, string) (domain.Attestation, error)
}
type Memory struct {
	mu       sync.RWMutex
	projects map[string]domain.Project
	defs     map[string]domain.BuildDefinition
	execs    map[string]domain.Execution
	workers  map[string]domain.Worker
	cache    map[digest.Digest]domain.CacheEntry
	atts     map[string]domain.Attestation
}

func NewMemory() *Memory {
	return &Memory{projects: map[string]domain.Project{}, defs: map[string]domain.BuildDefinition{}, execs: map[string]domain.Execution{}, workers: map[string]domain.Worker{}, cache: map[digest.Digest]domain.CacheEntry{}, atts: map[string]domain.Attestation{}}
}
func (m *Memory) CreateProject(_ context.Context, p domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[p.ID]; ok {
		return domain.ErrConflict
	}
	m.projects[p.ID] = p
	return nil
}
func (m *Memory) GetProject(_ context.Context, id string) (domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.projects[id]
	if !ok {
		return p, domain.ErrNotFound
	}
	return p, nil
}
func (m *Memory) SaveDefinition(_ context.Context, d domain.BuildDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defs[d.ID] = d
	return nil
}
func (m *Memory) GetDefinition(_ context.Context, id string) (domain.BuildDefinition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.defs[id]
	if !ok {
		return d, fmt.Errorf("definition %s missing: %w", id, domain.ErrNotFound)
	}
	return d, nil
}
func (m *Memory) SaveExecution(_ context.Context, e domain.Execution) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execs[e.ID] = e
	return nil
}
func (m *Memory) GetExecution(_ context.Context, id string) (domain.Execution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.execs[id]
	if !ok {
		return e, fmt.Errorf("execution %s missing: %w", id, domain.ErrNotFound)
	}
	return e, nil
}
func (m *Memory) FindExecutionByIdempotency(_ context.Context, key string) (domain.Execution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.execs {
		if e.IdempotencyKey == key {
			return e, nil
		}
	}
	return domain.Execution{}, domain.ErrNotFound
}
func (m *Memory) ListExecutions(_ context.Context) []domain.Execution {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Execution, 0, len(m.execs))
	for _, e := range m.execs {
		out = append(out, e)
	}
	return out
}
func (m *Memory) SaveWorker(_ context.Context, w domain.Worker) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workers[w.ID] = w
	return nil
}
func (m *Memory) GetWorker(_ context.Context, id string) (domain.Worker, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.workers[id]
	if !ok {
		return w, domain.ErrNotFound
	}
	return w, nil
}
func (m *Memory) ListWorkers(_ context.Context) []domain.Worker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Worker, 0, len(m.workers))
	for _, w := range m.workers {
		out = append(out, w)
	}
	return out
}
func (m *Memory) PutCache(_ context.Context, c domain.CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache[c.ActionKey] = c
	return nil
}
func (m *Memory) GetCache(_ context.Context, k digest.Digest) (domain.CacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.cache[k]
	if !ok || (!c.ExpiresAt.IsZero() && time.Now().After(c.ExpiresAt)) {
		return c, domain.ErrNotFound
	}
	return c, nil
}
func (m *Memory) SaveAttestation(_ context.Context, a domain.Attestation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.atts[a.ID] = a
	return nil
}
func (m *Memory) GetAttestation(_ context.Context, id string) (domain.Attestation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.atts[id]
	if !ok {
		return a, domain.ErrNotFound
	}
	return a, nil
}
