package datautils

import "sync"

type Cache[K comparable, V any] struct {
	registry map[K]V
	mu       sync.RWMutex
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		registry: make(map[K]V),
	}
}

func (c *Cache[K, V]) Get(name K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.registry[name]

	return val, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.registry[key] = value
}
