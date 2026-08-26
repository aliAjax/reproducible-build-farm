package infrastructure

import (
	"context"
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestCanceledContextStopsExecution(t *testing.T) {
	e := NewSimulatedExecutor()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.Execute(ctx, domain.Step{ID: "s1"}, nil)
	if err == nil {
		t.Fatal("executor ran a step with a canceled context")
	}
	if err != context.Canceled && ctx.Err() != context.Canceled {
		t.Fatalf("expected cancellation error, got %v", err)
	}
}
