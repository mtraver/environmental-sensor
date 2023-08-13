package cache

import (
	"sync"
	"time"
)

type Cache[V any] struct {
	mu      sync.Mutex
	entries map[string]entry[V]
}

type entry[V any] struct {
	value V
	exp   time.Time
}

func New[V any]() *Cache[V] {
	return &Cache[V]{
		entries: make(map[string]entry[V]),
	}
}

func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
	e := entry[V]{
		value: value,
		exp:   time.Now().Add(ttl),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = e
}

func (c *Cache[V]) Get(key string) V {
	c.mu.Lock()
	defer c.mu.Unlock()

	var val V

	e, ok := c.entries[key]
	if !ok {
		return val
	}

	// Present and unexpired
	if time.Now().Before(e.exp) {
		return e.value
	}

	// Expired
	delete(c.entries, key)
	return val
}

func (c *Cache[V]) clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	toRemove := []string{}

	for k, e := range c.entries {
		if !time.Now().Before(e.exp) {
			toRemove = append(toRemove, k)
		}
	}

	for _, k := range toRemove {
		delete(c.entries, k)
	}
}
