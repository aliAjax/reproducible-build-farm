package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"example.com/reproducible-build-farm/internal/domain"
	"example.com/reproducible-build-farm/pkg/digest"
)

func key(n string) digest.Digest { return digest.OfString(n) }

func TestConcurrentExpiredGetNoRace(t *testing.T) {
	c := NewMemory(10)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 50; j++ {
				_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("ea"), ExpiresAt: time.Now().Add(-time.Minute)})
				_, _ = c.Get(context.Background(), key("ea"))
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestPutEvictsOldestNotNewest(t *testing.T) {
	c := NewMemory(2)
	now := time.Now()
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("a"), CreatedAt: now.Add(-2 * time.Minute), ExpiresAt: now.Add(time.Hour)})
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("b"), CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)})
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("c"), CreatedAt: now, ExpiresAt: now.Add(time.Hour)})
	if _, ok := c.Get(context.Background(), key("a")); ok {
		t.Fatal("oldest entry must be evicted when cache is full")
	}
	if _, ok := c.Get(context.Background(), key("b")); !ok {
		t.Fatal("newest entry must survive eviction")
	}
	if _, ok := c.Get(context.Background(), key("c")); !ok {
		t.Fatal("newly added entry must be present")
	}
}

func TestPutEvictsExpiredFirst(t *testing.T) {
	c := NewMemory(2)
	now := time.Now()
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("exp"), CreatedAt: now.Add(-2 * time.Minute), ExpiresAt: now.Add(-time.Minute)})
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("live"), CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)})
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("new"), CreatedAt: now, ExpiresAt: now.Add(time.Hour)})
	if _, ok := c.Get(context.Background(), key("exp")); ok {
		t.Fatal("expired entry must be evicted first")
	}
	if _, ok := c.Get(context.Background(), key("live")); !ok {
		t.Fatal("live entry must survive when an expired one exists")
	}
	if _, ok := c.Get(context.Background(), key("new")); !ok {
		t.Fatal("newly added entry must be present")
	}
}

func TestGetRemovesExpired(t *testing.T) {
	c := NewMemory(2)
	now := time.Now()
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("exp"), CreatedAt: now.Add(-2 * time.Minute), ExpiresAt: now.Add(-time.Minute)})
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("live"), CreatedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)})
	if _, ok := c.Get(context.Background(), key("exp")); ok {
		t.Fatal("expired entry must be a miss")
	}
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("new"), CreatedAt: now, ExpiresAt: now.Add(time.Hour)})
	if _, ok := c.Get(context.Background(), key("live")); !ok {
		t.Fatal("live entry must survive after expired entry was lazily removed")
	}
	if _, ok := c.Get(context.Background(), key("new")); !ok {
		t.Fatal("new entry must fit without evicting live entries")
	}
}

func TestPutDefaultsTTL(t *testing.T) {
	c := NewMemory(10)
	_ = c.Put(context.Background(), domain.CacheEntry{ActionKey: key("a"), CreatedAt: time.Now()})
	v, ok := c.Get(context.Background(), key("a"))
	if !ok {
		t.Fatal("entry missing")
	}
	if v.ExpiresAt.IsZero() {
		t.Fatal("entry without explicit TTL must get a default expiry")
	}
}
