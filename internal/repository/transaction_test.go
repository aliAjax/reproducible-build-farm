package repository

import (
	"context"
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestCommitClearsStaged(t *testing.T) {
	m := NewMemory()
	tx := m.Begin(context.Background())
	for i := 0; i < 2; i++ {
		e := domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateQueued}
		if i == 1 {
			e.ID = "ex-2"
		}
		if err := tx.StageExecution(e); err != nil {
			t.Fatal(err)
		}
	}
	if tx.Size() != 2 {
		t.Fatalf("expected 2 staged, got %d", tx.Size())
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if tx.Size() != 0 {
		t.Fatalf("committed transaction must clear staged executions, got %d staged", tx.Size())
	}
}

func TestCommitRejectsDoubleCommit(t *testing.T) {
	m := NewMemory()
	tx := m.Begin(context.Background())
	if err := tx.StageExecution(domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateQueued}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); err == nil {
		t.Fatal("second commit must be rejected")
	}
}

func TestRollbackMarksCommitted(t *testing.T) {
	m := NewMemory()
	tx := m.Begin(context.Background())
	if err := tx.StageExecution(domain.Execution{ID: "ex-1", DefinitionID: "d", State: domain.StateQueued}); err != nil {
		t.Fatal(err)
	}
	tx.Rollback()
	if err := tx.Commit(context.Background()); err == nil {
		t.Fatal("commit after rollback must be rejected")
	}
}
