package accessibility

import (
	"sync"
	"time"
	"unsafe"

	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
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
		logger.Debug("Cache miss - element not found in cache", zap.Uintptr("element_ptr", key))
		return nil
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		logger.Debug("Cache miss - element expired",
			zap.Uintptr("element_ptr", key),
			zap.Time("expires_at", cached.ExpiresAt),
			zap.Time("current_time", time.Now()))
		return nil
	}

	logger.Debug("Cache hit",
		zap.Uintptr("element_ptr", key),
		zap.Time("expires_at", cached.ExpiresAt))
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

	logger.Debug("Cached element info",
		zap.Uintptr("element_ptr", key),
		zap.String("role", info.Role),
		zap.String("title", info.Title),
		zap.Time("expires_at", time.Now().Add(c.ttl)))
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
			logger.Debug("Cache cleanup loop stopped")
			return
		}
	}
}

// cleanup removes expired entries
func (c *InfoCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removedCount := 0
	for key, cached := range c.data {
		if now.After(cached.ExpiresAt) {
			delete(c.data, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		logger.Debug("Cache cleanup completed",
			zap.Int("removed_entries", removedCount),
			zap.Int("remaining_entries", len(c.data)))
	}
}

// Stop stops the cleanup goroutine
func (c *InfoCache) Stop() {
	if !c.stopped {
		close(c.stopCh)
		c.stopped = true
		logger.Debug("Cache stopped")
	}
}

// Clear removes all entries
func (c *InfoCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[uintptr]*CachedInfo, 100)
	logger.Debug("Cache cleared")
}

// Size returns the number of cached entries
func (c *InfoCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	size := len(c.data)
	logger.Debug("Cache size queried", zap.Int("size", size))
	return size
}
