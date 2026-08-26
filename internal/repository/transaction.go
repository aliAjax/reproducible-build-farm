package repository

import (
	"context"
	"example.com/reproducible-build-farm/internal/domain"
	"fmt"
	"time"
)

type Transaction struct {
	store     *Memory
	staged    []domain.Execution
	committed bool
}

func (m *Memory) Begin(_ context.Context) *Transaction {
	return &Transaction{store: m, staged: []domain.Execution{}}
}
func (t *Transaction) StageExecution(e domain.Execution) error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	if err := e.Validate(); err != nil {
		return err
	}
	t.staged = append(t.staged, e)
	return nil
}
func (t *Transaction) Commit(ctx context.Context) error {
	if t.committed {
		return fmt.Errorf("transaction already committed")
	}
	for _, e := range t.staged {
		if err := t.store.SaveExecution(ctx, e); err != nil {
			return err
		}
	}
	t.committed = true
	return nil
}
func (t *Transaction) Rollback() { t.staged = nil; t.committed = true }
func (t *Transaction) Size() int { return len(t.staged) }
func (m *Memory) TransitionExecution(ctx context.Context, id string, to domain.ExecutionState, reason string) (domain.Execution, error) {
	e, err := m.GetExecution(ctx, id)
	if err != nil {
		return e, err
	}
	if !domain.CanTransition(e.State, to) {
		return e, fmt.Errorf("cannot transition %s to %s", e.State, to)
	}
	e.State = to
	e.Error = reason
	if to == domain.StateRunning {
		e.StartedAt = time.Now().UTC()
	}
	if e.Terminal() {
		e.FinishedAt = time.Now().UTC()
	}
	return e, m.SaveExecution(ctx, e)
}
func (m *Memory) CountByState(ctx context.Context) map[domain.ExecutionState]int {
	out := map[domain.ExecutionState]int{}
	for _, e := range m.ListExecutions(ctx) {
		out[e.State]++
	}
	return out
}
func (m *Memory) CloneExecution(ctx context.Context, id string) (domain.Execution, error) {
	e, err := m.GetExecution(ctx, id)
	if err != nil {
		return e, err
	}
	e.ID = e.ID + "-retry"
	e.State = domain.StateQueued
	e.Attempt++
	e.Error = ""
	e.AttestationID = ""
	e.ResultDigest = ""
	e.CreatedAt = time.Now().UTC()
	e.StartedAt = time.Time{}
	e.FinishedAt = time.Time{}
	if err = m.SaveExecution(ctx, e); err != nil {
		return e, err
	}
	return e, nil
}
