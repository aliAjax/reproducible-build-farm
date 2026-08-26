package observability

import (
	"fmt"
	"sync"
	"time"
)

type Counter struct {
	mu     sync.Mutex
	values map[string]int64
}

func NewCounter() *Counter               { return &Counter{values: map[string]int64{}} }
func (c *Counter) Inc(name string)       { c.mu.Lock(); defer c.mu.Unlock(); c.values[name]++ }
func (c *Counter) Get(name string) int64 { c.mu.Lock(); defer c.mu.Unlock(); return c.values[name] }
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := map[string]int64{}
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

type Timer struct{ start time.Time }

func StartTimer() Timer                 { return Timer{start: time.Now()} }
func (t Timer) Duration() time.Duration { return time.Since(t.start) }
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dus", d.Microseconds())
	}
	return d.Round(time.Millisecond).String()
}
