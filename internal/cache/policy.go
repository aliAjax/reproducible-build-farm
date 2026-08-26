package cache

import (
	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
	"fmt"
	"strings"
	"time"
)

type Policy struct {
	TTL              time.Duration
	MaxOutputBytes   int64
	AllowedPlatforms map[string]bool
}

func DefaultPolicy() Policy {
	return Policy{TTL: 24 * time.Hour, MaxOutputBytes: 1 << 30, AllowedPlatforms: map[string]bool{"linux/amd64": true, "linux/arm64": true}}
}
func (p Policy) Validate() error {
	if p.TTL <= 0 {
		return fmt.Errorf("cache ttl must be positive")
	}
	if p.MaxOutputBytes <= 0 {
		return fmt.Errorf("max output bytes must be positive")
	}
	if len(p.AllowedPlatforms) == 0 {
		return fmt.Errorf("at least one platform required")
	}
	return nil
}
func (p Policy) Accept(platform string, size int64) error {
	if !p.AllowedPlatforms[platform] {
		return fmt.Errorf("platform %s is not allowed", platform)
	}
	if size < 0 || size > p.MaxOutputBytes {
		return fmt.Errorf("output size exceeds policy")
	}
	return nil
}
func Negative(key digest.Digest, ttl time.Duration) domain.CacheEntry {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return domain.CacheEntry{ActionKey: key, Negative: true, CreatedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(ttl)}
}
func KeyPrefix(key digest.Digest, n int) string {
	if n < 1 {
		n = 8
	}
	if n > len(key) {
		n = len(key)
	}
	return strings.ToLower(key.String()[:n])
}
