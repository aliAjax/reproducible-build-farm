package repository

import (
	"context"
	"errors"
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestDefinitionMissingKeepsErrorChain(t *testing.T) {
	m := NewMemory()
	_, err := m.GetDefinition(context.Background(), "no-such-def")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound chain, got %v", err)
	}
}

func TestExecutionMissingKeepsErrorChain(t *testing.T) {
	m := NewMemory()
	_, err := m.GetExecution(context.Background(), "no-such-exec")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound chain, got %v", err)
	}
}
