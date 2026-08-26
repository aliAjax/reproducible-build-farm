package dsl

import (
	"testing"

	"example.com/reproducible-build-farm/internal/domain"
)

func TestCanonicalDoesNotMutateDefinition(t *testing.T) {
	doc := Document{
		Name:        "n",
		ToolchainID: "tc",
		Steps: []domain.Step{
			{ID: "s2", Args: []string{"b", "a"}},
			{ID: "s1"},
		},
	}
	_, err := Canonical(doc)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Steps[0].ID != "s2" || doc.Steps[1].ID != "s1" {
		t.Fatalf("Canonical mutated definition step order: %v %v", doc.Steps[0].ID, doc.Steps[1].ID)
	}
	if len(doc.Steps[0].Args) != 2 || doc.Steps[0].Args[0] != "b" || doc.Steps[0].Args[1] != "a" {
		t.Fatalf("Canonical mutated shared step Args: %v", doc.Steps[0].Args)
	}
}

func TestCanonicalStableTwice(t *testing.T) {
	mk := func() Document {
		return Document{Name: "n", ToolchainID: "tc", Steps: []domain.Step{
			{ID: "s2", Args: []string{"b", "a"}},
			{ID: "s1", Dependencies: []string{"d2", "d1"}},
		}}
	}
	d1 := mk()
	d2 := mk()
	b1, err := Canonical(d1)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := Canonical(d2)
	if err != nil {
		t.Fatal(err)
	}
	if string(b1) != string(b2) {
		t.Fatalf("canonical not stable across identical definitions: %s vs %s", b1, b2)
	}
}
