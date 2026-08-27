package worker

import (
	"context"
	"testing"
	"time"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
)

func TestRegisterMarksWorkerFresh(t *testing.T) {
	store := repository.NewMemory()
	r := NewRegistry(store)
	if err := r.Register(context.Background(), domain.Worker{ID: "w1", Platform: "linux/amd64"}); err != nil {
		t.Fatal(err)
	}
	healthy := r.Healthy(context.Background(), time.Now())
	if len(healthy) != 1 {
		t.Fatalf("freshly registered worker must be healthy, got %d", len(healthy))
	}
}

func TestHeartbeatRefreshesWorker(t *testing.T) {
	store := repository.NewMemory()
	r := NewRegistry(store)
	ctx := context.Background()
	if err := r.Register(ctx, domain.Worker{ID: "w1", Platform: "linux/amd64"}); err != nil {
		t.Fatal(err)
	}
	// simulate a stale heartbeat, then heartbeat again
	w, err := store.GetWorker(ctx, "w1")
	if err != nil {
		t.Fatal(err)
	}
	w.LastHeartbeat = time.Now().Add(-2 * time.Hour)
	if err := store.SaveWorker(ctx, w); err != nil {
		t.Fatal(err)
	}
	if err := r.Heartbeat(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	healthy := r.Healthy(ctx, time.Now())
	if len(healthy) != 1 {
		t.Fatalf("heartbeated worker must be healthy, got %d", len(healthy))
	}
}

func TestHealthyWindowAllowsDoubleHeartbeat(t *testing.T) {
	store := repository.NewMemory()
	r := NewRegistry(store)
	if err := r.Register(context.Background(), domain.Worker{ID: "w1", Platform: "linux/amd64"}); err != nil {
		t.Fatal(err)
	}
	w, err := store.GetWorker(context.Background(), "w1")
	if err != nil {
		t.Fatal(err)
	}
	// a worker that heartbeated ~20s ago (within 2x15s) must stay healthy
	healthy := r.Healthy(context.Background(), w.LastHeartbeat.Add(20*time.Second))
	if len(healthy) != 1 {
		t.Fatalf("worker within double heartbeat window must be healthy, got %d", len(healthy))
	}
}
