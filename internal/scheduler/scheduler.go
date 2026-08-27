package scheduler

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Queue struct {
	mu       sync.Mutex
	items    []domain.Execution
	store    repository.Store
	capacity int
}

func NewQueue(s repository.Store, capacity int) *Queue {
	if capacity < 1 {
		capacity = 1000
	}
	return &Queue{store: s, capacity: capacity}
}
func (q *Queue) Enqueue(ctx context.Context, e domain.Execution) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) >= q.capacity {
		return fmt.Errorf("queue full")
	}
	q.items = append(q.items, e)
	return q.store.SaveExecution(ctx, e)
}
func (q *Queue) Dequeue() (domain.Execution, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return domain.Execution{}, false
	}
	e := q.items[0]
	q.items = q.items[1:]
	return e, true
}
func (q *Queue) Snapshot() []domain.Execution {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := append([]domain.Execution(nil), q.items...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func (q *Queue) RequeueExpired(ctx context.Context, now time.Time) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := 0
	for _, e := range q.items {
		if e.State == domain.StateRunning && now.Sub(e.StartedAt) > time.Hour {
			e.State = domain.StateQueued
			q.items = append(q.items, e)
			_ = q.store.SaveExecution(ctx, e)
			n++
		}
	}
	return n
}
