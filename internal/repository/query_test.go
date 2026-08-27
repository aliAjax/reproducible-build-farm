package repository

import (
	"context"
	"testing"
	"time"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
)

func TestDeleteExpiredCacheRemovesAll(t *testing.T) {
	m := NewMemory()
	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		_ = m.PutCache(context.Background(), domain.CacheEntry{ActionKey: digest.OfString("exp-" + string(rune('a'+i))), ExpiresAt: time.Unix(now-10, 0)})
	}
	_ = m.PutCache(context.Background(), domain.CacheEntry{ActionKey: digest.OfString("live"), ExpiresAt: time.Unix(now+3600, 0)})
	n := m.DeleteExpiredCache(now)
	if n != 3 {
		t.Fatalf("expected 3 expired deleted, got %d", n)
	}
	if _, err := m.GetCache(context.Background(), digest.OfString("live")); err != nil {
		t.Fatal("live cache entry must survive cleanup")
	}
	if _, err := m.GetCache(context.Background(), digest.OfString("exp-a")); err == nil {
		t.Fatal("expired cache entry must be removed")
	}
}
