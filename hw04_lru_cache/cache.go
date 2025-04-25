package hw04lrucache

import "sync"

type Key string

type cacheItem struct {
	key   Key
	value interface{}
	// Храним ключи иначе не удалить будет
}

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mut      sync.Mutex
}

// Clear implements Cache.
func (c *lruCache) Clear() {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.items = make(map[Key]*ListItem, c.capacity)
	c.queue = NewList()
}

// Get implements Cache.
func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item)
		cacheItem := item.Value.(*cacheItem)
		return cacheItem.value, true
	}
	return nil, false
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mut.Lock()
	defer c.mut.Unlock()
	if item, ok := c.items[key]; ok {
		item.Value.(*cacheItem).value = value
		c.queue.MoveToFront(item)
		return true
	}
	newItem := c.queue.PushFront(&cacheItem{key: key, value: value})
	c.items[key] = newItem
	if c.capacity < c.queue.Len() {
		delete(c.items, c.queue.Back().Value.(*cacheItem).key)
		c.queue.Remove(c.queue.Back())
	}
	return false
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
