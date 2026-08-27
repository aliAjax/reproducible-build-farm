package scheduler

import (
	"context"
	"testing"
	"time"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/internal/repository"
)

func TestEnqueueRejectsDuplicate(t *testing.T) {
	store := repository.NewMemory()
	q := NewQueue(store, 100)
	e := domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateQueued}
	if err := q.Enqueue(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue(context.Background(), e); err == nil {
		t.Fatal("duplicate enqueue must be rejected")
	}
}

func TestRequeueExpiredNoDuplicate(t *testing.T) {
	store := repository.NewMemory()
	q := NewQueue(store, 100)
	e := domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateRunning, StartedAt: time.Now().Add(-2 * time.Hour)}
	if err := q.Enqueue(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	n := q.RequeueExpired(context.Background(), time.Now())
	if n != 1 {
		t.Fatalf("expected 1 requeued, got %d", n)
	}
	if len(q.Snapshot()) != 1 {
		t.Fatalf("requeue must not duplicate the execution, got %d items", len(q.Snapshot()))
	}
	if q.Snapshot()[0].State != domain.StateQueued {
		t.Fatalf("requeued execution must be queued, got %s", q.Snapshot()[0].State)
	}
}

func TestRequeueExpiredIncrementsAttempt(t *testing.T) {
	store := repository.NewMemory()
	q := NewQueue(store, 100)
	e := domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateRunning, StartedAt: time.Now().Add(-2 * time.Hour)}
	if err := q.Enqueue(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	q.RequeueExpired(context.Background(), time.Now())
	got, err := store.GetExecution(context.Background(), "ex-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Attempt != 1 {
		t.Fatalf("requeue must increment attempt, got %d", got.Attempt)
	}
}

func TestRequeueExpiredMarksFailedAfterLimit(t *testing.T) {
	store := repository.NewMemory()
	q := NewQueue(store, 100)
	e := domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateRunning, Attempt: 3, StartedAt: time.Now().Add(-2 * time.Hour)}
	if err := q.Enqueue(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	q.RequeueExpired(context.Background(), time.Now())
	got, err := store.GetExecution(context.Background(), "ex-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != domain.StateFailed {
		t.Fatalf("execution past retry limit must be failed, got %s", got.State)
	}
}
