package clock

import (
	"context"
	"time"
)

func WithTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		d = time.Minute
	}
	return context.WithTimeout(parent, d)
}
func DeadlineExceeded(err error) bool { return err == context.DeadlineExceeded }
