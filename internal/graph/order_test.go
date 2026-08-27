package graph

import (
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestOrderRejectsCycle(t *testing.T) {
	steps := []domain.Step{
		{ID: "a", Dependencies: []string{"b"}},
		{ID: "b", Dependencies: []string{"a"}},
	}
	if _, err := Order(steps); err == nil {
		t.Fatal("cyclic graph must be rejected by Order")
	}
}
