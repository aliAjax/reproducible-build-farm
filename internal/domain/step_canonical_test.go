package domain

import "testing"

func TestStepCanonicalDoesNotMutateArgs(t *testing.T) {
	s := Step{ID: "s1", Args: []string{"build", "a", "z"}}
	_ = s.Canonical()
	if len(s.Args) != 3 || s.Args[0] != "build" || s.Args[1] != "a" || s.Args[2] != "z" {
		t.Fatalf("Canonical mutated shared Args: %v", s.Args)
	}
}

func TestStepCanonicalDoesNotMutateDeps(t *testing.T) {
	s := Step{ID: "s1", Dependencies: []string{"d2", "d1"}}
	_ = s.Canonical()
	if len(s.Dependencies) != 2 || s.Dependencies[0] != "d2" || s.Dependencies[1] != "d1" {
		t.Fatalf("Canonical mutated shared Dependencies: %v", s.Dependencies)
	}
}
