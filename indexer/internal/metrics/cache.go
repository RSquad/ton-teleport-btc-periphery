package metrics

import (
	"time"
)

type CacheItem[T any] struct {
	Value      T
	Expiration int64
}

type Cache[T any] struct {
	items map[string]CacheItem[T]
	//mu    sync.RWMutex
}

func NewCache[T any]() *Cache[T] {
	cache := &Cache[T]{
		items: make(map[string]CacheItem[T]),
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

func (c *Cache[T]) Set(key string, value T, duration time.Duration) {
	//c.mu.Lock()
	//defer c.mu.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = CacheItem[T]{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	//c.mu.RLock()
	//defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, false
	}

	// Check if the item has expired
	if time.Now().UnixNano() > item.Expiration {
		var zero T
		return zero, false
	}

	return item.Value, true
}

func (c *Cache[T]) DeleteExpired() {
	//c.mu.Lock()
	//defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range c.items {
		if now > v.Expiration {
			delete(c.items, k)
		}
	}
}
