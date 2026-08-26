package audit

import "testing"

func TestAuditListIsolation(t *testing.T) {
	l := New()
	l.Append("u-1", "run", "r-1", map[string]string{"k": "v"})
	got := l.List()
	got[0].Metadata["k"] = "changed"
	got2 := l.List()
	if got2[0].Metadata["k"] != "v" {
		t.Fatalf("audit list shares mutable metadata, got %q", got2[0].Metadata["k"])
	}
}
