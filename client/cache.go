package client

import "time"

type cacheEntry[T any] struct {
	value  T
	expiry time.Time
}

type simpleCache[T any] struct {
	items map[string]cacheEntry[T]
}

func newSimpleCache[T any]() *simpleCache[T] {
	return &simpleCache[T]{items: make(map[string]cacheEntry[T])}
}

func (c *simpleCache[T]) get(key string) (T, bool) {
	if c == nil {
		var zero T
		return zero, false
	}

	entry, ok := c.items[key]
	if !ok {
		var zero T
		return zero, false
	}

	if entry.expiry.IsZero() || time.Now().After(entry.expiry) {
		delete(c.items, key)
		var zero T
		return zero, false
	}

	return entry.value, true
}

func (c *simpleCache[T]) set(key string, value T, expiry time.Time) {
	if c == nil {
		return
	}

	c.items[key] = cacheEntry[T]{value: value, expiry: expiry}
}

func (c *simpleCache[T]) clear() {
	if c == nil {
		return
	}
	c.items = make(map[string]cacheEntry[T])
}
