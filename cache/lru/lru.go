// Package lru provides a basic least-recently-used cache.
//
// It supports per-entry cost, and a custom on-evict callback.
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

// Cost is a measure of how much an entry "costs".
//
// The cache is limited to a certain maximum total cost.
type Cost int64

// cell is the type actually stored in each list entry.
type cell struct {
	key   Key
	value interface{}
	cost  Cost
}

// RetrieverFunc is called when the cache is missing a necessary value.
//
// If it returns an error, the value is not added to the cache, and the error
// is returned from `Get`.
type RetrieverFunc func(key Key) (value interface{}, cost Cost, err error)

// EvictionFunc is called when the cache evicts a value.
type EvictionFunc func(key Key, value interface{})

// Cache is the main cache type.
//
// Not internally synchronized.
type Cache struct {
	list    *list.List            // Entries are `cell` structs.
	entries map[Key]*list.Element // Same entries as in the list.
	cost    Cost

	// MaxCost is the cost of entries allowed in the cache.
	//
	// If reduced between calls to Get, the next call to Get that adjusts the
	// contents of the cache will reduce the cache size.
	//
	// The total cost of the cache may be higher than this, but only if due to
	// a single "jumbo" entry whose cost is greater than this.
	MaxCost Cost

	retriever RetrieverFunc

	// OnEvict, if not nil, is called each time a cache entry is evicted.
	OnEvict EvictionFunc
}

// New returns a new LRU cache with the given maximum size and retriever
// function.
//
// If you want to limit strictly to entry count, set the `maxCost` to the
// desired maximum number of entries, and return a cost of 1 from your
// retriever function.
func New(maxCost Cost, retriever RetrieverFunc) *Cache {
	return &Cache{
		list:      list.New(),
		entries:   make(map[Key]*list.Element),
		MaxCost:   maxCost,
		retriever: retriever,
	}
}

// Get retrieves an entry.
//
// If necessary, the cache will request the entry from the RetrieverFunc.
func (c *Cache) Get(key Key) (value interface{}, err error) {
	entry, ok := c.entries[key]
	if ok {
		c.list.MoveToBack(entry)
		return entry.Value.(cell).value, nil
	}

	var cost Cost
	value, cost, err = c.retriever(key)
	if err != nil {
		return
	}
	c.entries[key] = c.list.PushBack(cell{key, value, cost})
	c.cost += cost

	for c.cost > c.MaxCost && len(c.entries) > 1 {
		c.Evict()
	}
	return
}

// Clear evicts every entry in the cache.
func (c *Cache) Clear() {
	for len(c.entries) > 0 {
		c.Evict()
	}
}

// Evict evicts the least recently used entry from the cache.
func (c *Cache) Evict() {
	value := c.list.Remove(c.list.Front()).(cell)
	delete(c.entries, value.key)
	c.cost -= value.cost
	if c.OnEvict != nil {
		c.OnEvict(value.key, value.value)
	}
}
