package worker

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Registry struct {
	store     repository.Store
	heartbeat time.Duration
}

func NewRegistry(s repository.Store) *Registry {
	return &Registry{store: s, heartbeat: 15 * time.Second}
}
func (r *Registry) Register(ctx context.Context, w domain.Worker) error {
	if w.ID == "" || w.Platform == "" {
		return fmt.Errorf("worker id and platform are required")
	}
	w.LastHeartbeat = time.Now().UTC()
	w.Busy = false
	return r.store.SaveWorker(ctx, w)
}
func (r *Registry) Heartbeat(ctx context.Context, id string) error {
	w, e := r.store.GetWorker(ctx, id)
	if e != nil {
		return e
	}
	w.LastHeartbeat = time.Now().UTC()
	return r.store.SaveWorker(ctx, w)
}
func (r *Registry) Healthy(ctx context.Context, now time.Time) []domain.Worker {
	all := r.store.ListWorkers(ctx)
	out := []domain.Worker{}
	for _, w := range all {
		if now.Sub(w.LastHeartbeat) <= 2*r.heartbeat {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.Compare(out[i].ID, out[j].ID) < 0 })
	return out
}
func (r *Registry) Drain(ctx context.Context, id string) error {
	w, e := r.store.GetWorker(ctx, id)
	if e != nil {
		return e
	}
	w.Capacity = domain.ResourceBudget{}
	return r.store.SaveWorker(ctx, w)
}
