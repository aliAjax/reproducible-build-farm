package worker

import (
	"context"
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
)

func TestAcquireRespectsCanceledContext(t *testing.T) {
	store := repository.NewMemory()
	_ = store.SaveWorker(context.Background(), domain.Worker{ID: "w1", Platform: "linux/amd64", Version: "1", Capacity: domain.ResourceBudget{CPU: 8, MemoryMB: 8192}})
	m := NewManager(store, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.Acquire(ctx, "ex-1", domain.ResourceBudget{CPU: 1, MemoryMB: 128})
	if err == nil {
		t.Fatal("Acquire must not grant a lease on a canceled context")
	}
}
