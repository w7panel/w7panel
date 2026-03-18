package helper

import (
	"sync"
	"time"
)

var (
	defaultCache = NewMemoryCache()
)

func Set(key string, value interface{}, duration time.Duration) {
	defaultCache.Set(key, value, duration)
}
func Get(key string) (interface{}, bool) {
	return defaultCache.Get(key)
}

func Check(key string, value interface{}) bool {
	val, ok := defaultCache.Get(key)
	return ok && val == value
}

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      interface{}
	Expiration int64 // Unix timestamp, 0 means no expiration
}

// MemoryCache is an in-memory cache with thread-safe operations
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheItem
}

// NewMemoryCache creates a new MemoryCache instance
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]*CacheItem),
	}
	// Start cleanup goroutine
	go c.cleanupExpired()
	return c
}

// Set stores a value in the cache with optional expiration
// duration of 0 means no expiration
func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value from the cache
// Returns the value and a boolean indicating if the key was found and not expired
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete removes a key from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Has checks if a key exists in the cache and is not expired
func (c *MemoryCache) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// Check if expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return false
	}

	return true
}

// Clear removes all items from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Count returns the number of items in the cache (including expired ones)
func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// cleanupExpired periodically removes expired items from the cache
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if item.Expiration > 0 && now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
