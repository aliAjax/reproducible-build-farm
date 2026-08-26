package observability

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestSnapshotIsolation(t *testing.T) {
	c := NewCounter()
	c.Inc("builds")
	c.Inc("builds")
	snap := c.Snapshot()
	snap["builds"] = 999
	if got := c.Get("builds"); got != 2 {
		t.Fatalf("snapshot mutation polluted the counter, got %d want 2", got)
	}
}

func TestLoggerFieldsIsolation(t *testing.T) {
	var buf bytes.Buffer
	base := NewLogger()
	base.out = log.New(&buf, "", 0)
	_ = base.With("tenant", "t-1")
	base.Info("hello", nil)
	if strings.Contains(buf.String(), "tenant") {
		t.Fatalf("child logger fields leaked into base logger: %s", buf.String())
	}
}
