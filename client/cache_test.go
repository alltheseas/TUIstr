package client

import (
	"testing"
	"time"
)

func TestSimpleCacheStoresAndExpires(t *testing.T) {
	cache := newSimpleCache[int]()

	cache.set("foo", 42, time.Now().Add(time.Second))
	if val, ok := cache.get("foo"); !ok || val != 42 {
		t.Fatalf("expected cached value 42, got %v ok=%v", val, ok)
	}

	cache.set("soon", 7, time.Now().Add(-time.Second))
	if _, ok := cache.get("soon"); ok {
		t.Fatalf("expected expired entry to be evicted")
	}
}

func TestSimpleCacheClear(t *testing.T) {
	cache := newSimpleCache[string]()
	cache.set("k", "v", time.Now().Add(time.Minute))
	cache.clear()
	if _, ok := cache.get("k"); ok {
		t.Fatalf("expected cache to be cleared")
	}
}
