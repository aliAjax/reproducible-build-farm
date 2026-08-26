package digest

import "testing"

func TestTreeLookupEmptySafe(t *testing.T) {
	var tr Tree
	e, ok := tr.Lookup("anything")
	if ok {
		t.Fatal("empty tree must not report a hit")
	}
	if e != (Entry{}) {
		t.Fatalf("empty tree lookup must return zero entry, got %+v", e)
	}
}

func TestTreeLookupMissingReturnsZero(t *testing.T) {
	tr := NewTree([]Entry{{Name: "a", Digest: OfString("1"), Size: 1}})
	e, ok := tr.Lookup("missing")
	if ok {
		t.Fatal("missing entry must not report a hit")
	}
	if e != (Entry{}) {
		t.Fatalf("missing lookup must return zero entry, got %+v", e)
	}
}
