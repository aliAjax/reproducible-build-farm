package audit

import "testing"

func TestAppendCopiesMetadata(t *testing.T) {
	l := New()
	meta := map[string]string{"k": "v"}
	l.Append("u-1", "run", "r-1", meta)
	meta["k"] = "changed"
	got := l.List()
	if got[0].Metadata["k"] != "v" {
		t.Fatalf("Append stored caller's metadata by reference, got %q", got[0].Metadata["k"])
	}
}
