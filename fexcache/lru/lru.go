package lru

import (
	"container/list"
)

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func (e *entry) Size() int64 { return int64(len(e.key) + e.value.Len()) }

type Cache struct {
	cache     map[string]*list.Element
	lst       *list.List
	maxBytes  int64
	nBytes    int64
	OnEvicted func(string, Value) // 删除时的回调函数
}

func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		cache:     make(map[string]*list.Element),
		maxBytes:  maxBytes,
		OnEvicted: onEvicted,
		lst:       list.New(),
	}
}

func (c *Cache) Len() int { return c.lst.Len() }

func (c *Cache) Get(key string) (Value, bool) {
	element, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	c.lst.MoveToFront(element)
	kv := element.Value.(*entry)
	return kv.value, true
}

func (c *Cache) prepareSpace(require int64) bool {
	for require+c.nBytes > c.maxBytes {
		if c.nBytes == 0 {
			return false
		}
		c.RemoveOldest()
	}
	return true
}

// Add 已存在则更新，不存在则新增。依据内存限制，先删后加
func (c *Cache) Add(key string, value Value) {
	element, ok := c.cache[key]
	if ok {
		kv := element.Value.(*entry)
		expBytes := int64(value.Len() - kv.value.Len())
		c.prepareSpace(expBytes)
		kv.value = value
		c.nBytes += expBytes
		c.lst.MoveToFront(element)
		return
	}
	newKV := &entry{key: key, value: value}
	c.prepareSpace(newKV.Size())
	element = c.lst.PushFront(newKV)
	c.nBytes += newKV.Size()
	c.cache[key] = element
}

func (c *Cache) RemoveOldest() {
	element := c.lst.Back()
	if element == nil {
		return
	}
	c.lst.Remove(element)
	kv := element.Value.(*entry)
	delete(c.cache, kv.key)
	c.nBytes -= kv.Size()
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}
