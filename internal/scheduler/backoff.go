package scheduler

import (
	"math"
	"math/rand"
	"time"
)

func Delay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := float64(base) * math.Pow(2, float64(attempt))
	if d > float64(max) {
		d = float64(max)
	}
	j := 0.8 + rand.New(rand.NewSource(int64(attempt)+1)).Float64()*0.4
	return time.Duration(d * j)
}
func ShouldRetry(attempt, limit int) bool { return attempt < limit }
