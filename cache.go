package fexcache

import (
	"sync"

	"github.com/felixorbit/fexcache/lru"
)

// 通过互斥锁进行并发访问控制
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.NewCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}
