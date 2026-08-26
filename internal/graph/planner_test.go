package graph

import (
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestBuildPlanDoesNotMutateInput(t *testing.T) {
	steps := []domain.Step{
		{ID: "s2", Dependencies: []string{"s1"}},
		{ID: "s1"},
	}
	_, err := BuildPlan(steps)
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].ID != "s2" || steps[1].ID != "s1" {
		t.Fatalf("BuildPlan mutated input step order: %v %v", steps[0].ID, steps[1].ID)
	}
}
