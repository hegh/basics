// Package lru provides a basic least-recently-used cache.
//
// Anticipated usage:
//
//	func retrieveEntry(key Key) (interface{}, error) {
//		// Expensive retrieval operation.
//	}
//	func evictEntry(key Key, value interface{}) {
//		// Optional release operation.
//	}
//	cache := lru.New(5, retrieveEntry)
//	cache.OnEvict = evictEntry
//	value, err := cache.Get(key)
//	cache.Clear()
package lru

import "container/list"

// Key can be any map-key-compatible type.
type Key interface{}

// pair is the type actually stored in each list entry.
type pair struct {
	key   Key
	value interface{}
}

// RetrieverFunc is called when the cache is missing a necessary value.
//
// If it returns an error, the value is not added to the cache, and the error
// is returned from `Get`.
type RetrieverFunc func(key Key) (value interface{}, err error)

// EvictionFunc is called when the cache evicts a value.
type EvictionFunc func(key Key, value interface{})

// Cache is the main cache type.
//
// Not internally synchronized.
type Cache struct {
	list    *list.List
	entries map[Key]*list.Element
	size    int

	// MaxEntries is the maximum number of entries allowed in the cache.
	// If reduced between calls to Get, the next call to Get that adjusts the
	// contents of the cache will reduce the cache size.
	MaxEntries int

	retriever RetrieverFunc

	// OnEvict, if not nil, is called each time a cache entry is evicted.
	OnEvict EvictionFunc
}

// New returns a new LRU cache with the given maximum size and retriever
// function.
func New(maxEntries int, retriever RetrieverFunc) *Cache {
	return &Cache{
		list:       list.New(),
		entries:    make(map[Key]*list.Element),
		MaxEntries: maxEntries,
		retriever:  retriever,
	}
}

// Get retrieves an entry.
//
// If necessary, the cache will request the entry from the RetrieverFunc.
func (c *Cache) Get(key Key) (value interface{}, err error) {
	entry, ok := c.entries[key]
	if ok {
		c.list.MoveToBack(entry)
		return entry.Value.(pair).value, nil
	}

	for c.size >= c.MaxEntries {
		c.evict()
	}

	value, err = c.retriever(key)
	if err != nil {
		return
	}
	c.entries[key] = c.list.PushBack(pair{key, value})
	c.size++
	return
}

// Clear evicts every entry in the cache.
func (c *Cache) Clear() {
	for c.size > 0 {
		c.evict()
	}
}

// evict evicts the least recently used entry from the cache.
func (c *Cache) evict() {
	value := c.list.Remove(c.list.Front()).(pair)
	delete(c.entries, value.key)
	c.size--
	if c.OnEvict != nil {
		c.OnEvict(value.key, value.value)
	}
}
