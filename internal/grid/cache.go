package grid

import (
    "container/list"
    "image"
    "sync"
    "time"
    "go.uber.org/zap"
)

type gridCacheKey struct {
    characters string
    width      int
    height     int
}

type gridCacheEntry struct {
    cells   []*Cell
    addedAt time.Time
    usedAt  time.Time
}

type GridCache struct {
    mu       sync.Mutex
    items    map[gridCacheKey]*list.Element
    order    *list.List
    capacity int
}

var gridCache = newGridCache(8)
var gridCacheEnabled = true

func SetGridCacheEnabled(enabled bool) {
    gridCacheEnabled = enabled
}

func Prewarm(characters string, sizes []image.Rectangle) {
    if !gridCacheEnabled {
        return
    }
    for _, r := range sizes {
        if _, ok := gridCache.get(characters, r); ok {
            continue
        }
        g := NewGrid(characters, r, zap.NewNop())
        _ = g
    }
}

func newGridCache(cap int) *GridCache {
    return &GridCache{items: make(map[gridCacheKey]*list.Element), order: list.New(), capacity: cap}
}

func (c *GridCache) get(characters string, bounds image.Rectangle) ([]*Cell, bool) {
    key := gridCacheKey{characters: characters, width: bounds.Dx(), height: bounds.Dy()}
    c.mu.Lock()
    defer c.mu.Unlock()
    if el, ok := c.items[key]; ok {
        ent := el.Value.(*gridCacheEntry)
        ent.usedAt = time.Now()
        c.order.MoveToFront(el)
        return ent.cells, true
    }
    return nil, false
}

func (c *GridCache) put(characters string, bounds image.Rectangle, cells []*Cell) {
    key := gridCacheKey{characters: characters, width: bounds.Dx(), height: bounds.Dy()}
    c.mu.Lock()
    defer c.mu.Unlock()
    if el, ok := c.items[key]; ok {
        ent := el.Value.(*gridCacheEntry)
        ent.cells = cells
        ent.usedAt = time.Now()
        ent.addedAt = ent.usedAt
        c.order.MoveToFront(el)
        return
    }
    ent := &gridCacheEntry{cells: cells, addedAt: time.Now(), usedAt: time.Now()}
    el := c.order.PushFront(ent)
    c.items[key] = el
    if c.order.Len() > c.capacity {
        tail := c.order.Back()
        if tail != nil {
            removed := tail.Value.(*gridCacheEntry)
            _ = removed
            c.order.Remove(tail)
            for k, v := range c.items {
                if v == tail {
                    delete(c.items, k)
                    break
                }
            }
        }
    }
}