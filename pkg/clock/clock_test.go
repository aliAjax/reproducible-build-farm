package clock

import (
	"testing"
	"time"
)

func TestFixedClockReturnsUTC(t *testing.T) {
	local := time.Date(2026, 8, 27, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	f := Fixed{T: local}
	got := f.Now()
	if got.Location() != time.UTC {
		t.Fatalf("Fixed clock must return UTC, got %v", got.Location())
	}
}
