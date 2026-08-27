package cache

import (
	"sync"
	"time"
)

// entry holds a cached response and its creation time.
type entry struct {
	data      []byte
	createdAt time.Time
}

// ResponseCache is a simple in-memory cache with a fixed TTL.
type ResponseCache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]entry
}

// New creates a ResponseCache with the given TTL.
func New(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		ttl:     ttl,
		entries: make(map[string]entry),
	}
}

// Get returns the cached data for the given key if it exists and has not expired.
func (c *ResponseCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Since(e.createdAt) > c.ttl {
		return nil, false
	}
	return e.data, true
}

// Set stores data under the given key, replacing any existing entry.
func (c *ResponseCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry{
		data:      data,
		createdAt: time.Now(),
	}
}
