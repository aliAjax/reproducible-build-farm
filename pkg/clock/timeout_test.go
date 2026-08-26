package clock

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWithTimeoutZeroUsesDefault(t *testing.T) {
	ctx, cancel := WithTimeout(context.Background(), 0)
	defer cancel()
	select {
	case <-ctx.Done():
		t.Fatal("zero timeout must fall back to the default window, not expire immediately")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestDeadlineExceededWrapped(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	<-ctx.Done()
	wrapped := fmt.Errorf("boom: %w", ctx.Err())
	if !DeadlineExceeded(wrapped) {
		t.Fatalf("wrapped deadline error not recognized: %v", wrapped)
	}
}
