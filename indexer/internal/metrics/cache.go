package metrics

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      string
	Expiration int64
}

type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	// Clear cache routine
	go func() {
		for {
			time.Sleep(time.Minute)
			cache.DeleteExpired()
		}
	}()

	return cache
}

func (c *Cache) Set(key, value string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return "", false
	}

	// Check if the item has expired
	if time.Now().UnixNano() > item.Expiration {
		return "", false
	}

	return item.Value, true
}

func (c *Cache) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range c.items {
		if now > v.Expiration {
			delete(c.items, k)
		}
	}
}
