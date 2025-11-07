package accessibility

import (
	"sync"
	"time"
	"unsafe"
)

// CachedInfo wraps ElementInfo with expiration time
type CachedInfo struct {
	Info      *ElementInfo
	ExpiresAt time.Time
}

// InfoCache is a thread-safe cache with TTL
type InfoCache struct {
	mu      sync.RWMutex
	data    map[uintptr]*CachedInfo
	ttl     time.Duration
	stopCh  chan struct{}
	stopped bool
}

// NewInfoCache creates a cache with the given TTL
func NewInfoCache(ttl time.Duration) *InfoCache {
	cache := &InfoCache{
		data:   make(map[uintptr]*CachedInfo, 100),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached value if it exists and hasn't expired
func (c *InfoCache) Get(elem *Element) *ElementInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := uintptr(unsafe.Pointer(elem))
	cached, exists := c.data[key]

	if !exists {
		return nil
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached.Info
}

// Set stores a value with TTL
func (c *InfoCache) Set(elem *Element, info *ElementInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := uintptr(unsafe.Pointer(elem))
	c.data[key] = &CachedInfo{
		Info:      info,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// cleanupLoop periodically removes expired entries
func (c *InfoCache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2) // Cleanup at half the TTL interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// cleanup removes expired entries
func (c *InfoCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, cached := range c.data {
		if now.After(cached.ExpiresAt) {
			delete(c.data, key)
		}
	}
}

// Stop stops the cleanup goroutine
func (c *InfoCache) Stop() {
	if !c.stopped {
		close(c.stopCh)
		c.stopped = true
	}
}

// Clear removes all entries
func (c *InfoCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[uintptr]*CachedInfo, 100)
}

// Size returns the number of cached entries
func (c *InfoCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
