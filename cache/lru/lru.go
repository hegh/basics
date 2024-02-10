// Package lru provides a basic least-recently-used cache. It can be used as
// either a read-through or a manual cache.
//
// It supports per-entry cost, and custom on-retrieve and on-evict callbacks.
//
// Anticipated usage (read-through):
//
//	func retrieveEntry(key lru.Key) (interface{}, lru.Cost, error) {
//		// Expensive retrieval operation.
//	}
//	func evictEntry(key lru.Key, value interface{}) {
//		// Optional release operation.
//	}
//	cache := lru.New(5)
//	cache.OnRetrieve = retrieveEntry
//	cache.OnEvict = evictEntry
//	value, err := cache.Get(key)
//	cache.Clear()
//
// Anticipated usage (manual caching):
//
//	cache := lru.New(5)
//	cache.Put(key1, value1)
//	cache.Put(key2, value2)
//	value, err := cache.Get(key1)
//	if err != nil {
//		// ... (must be ErrMissingEntry)
//	}
package lru

import (
	"container/list"
	"errors"
	"fmt"
	"math"
)

// ErrMissingEntry is returned from `Get` if there is no such entry in the
// cache, and there is no retriever function.
//
// If there is no retriever function, this is the only error than can be
// returned from `Get`.
var ErrMissingEntry = errors.New("missing entry")

// Key can be any map-key-compatible type.
type Key interface{}

// Cost is a measure of how much an entry "costs".
//
// The cache is limited to a chosen maximum total cost.
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
	list    *list.List            // Entries are `*cell`s.
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

	// OnRetrieve, if not nil, is called when Get does not find an entry in the
	// cache.
	//
	// If nil, the return value will be nil with a `ErrMissingEntry` error.
	OnRetrieve RetrieverFunc

	// OnEvict, if not nil, is called each time a cache entry is evicted.
	OnEvict EvictionFunc
}

// New returns a new LRU cache with the given maximum size.
//
// You may want to add a retriever and/or eviction function to the returned
// cache.
//
// If you want to limit by entry count, set the `maxCost` to the desired maximum
// number of entries, and return a cost of 1 from your retriever function.
//
// Entries with a cost of 0 cannot evict other entries, but they will themselves
// be evicted if something more expensive comes in and the 0-cost entries were
// the least recently used.
//
// Negative costs are not supported and will cause panics.
//
// Maximum cache cost is `math.MaxInt64`.
func New(maxCost Cost) *Cache {
	return &Cache{
		list:    list.New(),
		entries: make(map[Key]*list.Element),
		MaxCost: maxCost,
	}
}

// Cost returns the current cost of the entries in the cache.
func (c *Cache) Cost() Cost { return c.cost }

// Get retrieves an entry.
//
// If necessary and available, the cache will request the entry from the
// RetrieverFunc.
//
// Panics if the cost of a new entry would overflow the cache cost.
//
// If there is no retriever function, the only error that this can return is
// `ErrMissingEntry`. If there is a retriever function, this will return
// whatever error the retriever returned.
//
// If the retriever returns an error, the value will not be saved in the cache,
// but this will return whatever value the retriever returned.
func (c *Cache) Get(key Key) (value interface{}, err error) {
	entry, ok := c.entries[key]
	if ok {
		c.list.MoveToBack(entry)
		return entry.Value.(*cell).value, nil
	}

	if c.OnRetrieve == nil {
		return nil, ErrMissingEntry
	}

	var cost Cost
	value, cost, err = c.OnRetrieve(key)
	if err != nil {
		return
	}
	c.Put(key, cost, value)
	return
}

// Put directly adds an entry to the cache, or refreshes an existing entry.
//
// If the entry already existed in the cache, its cost and value will be
// updated to the values provided in this call.
//
// May cause evictions of other entries.
//
// Panics if the cost of the new entry would overflow the cache cost.
//
// Returns the previous value of the entry, or nil.
func (c *Cache) Put(key Key, cost Cost, value interface{}) interface{} {
	if cost < 0 {
		panic(fmt.Errorf("illegal cost: entry %v cost %d is negative", key, cost))
	}

	var prev interface{}
	entry, ok := c.entries[key]
	if ok {
		entryCell := entry.Value.(*cell)

		if c.cost-entryCell.cost+cost < 0 {
			panic(fmt.Errorf("cost overflow: cache cost %d + entry %v cost %d > limit %d", c.cost-entryCell.cost, key, cost, math.MaxInt64))
		}

		c.list.MoveToBack(entry)
		c.cost -= entryCell.cost
		c.cost += cost

		prev = entryCell.value
		entryCell.cost = cost
		entryCell.value = value
	} else {
		if c.cost+cost < 0 {
			panic(fmt.Errorf("cost overflow: cache cost %d + entry %v cost %d > limit %d", c.cost, key, cost, math.MaxInt64))
		}
		c.entries[key] = c.list.PushBack(&cell{key, value, cost})
		c.cost += cost
	}
	for c.cost > c.MaxCost && len(c.entries) > 1 {
		c.EvictOldest()
	}
	return prev
}

// Clear evicts every entry in the cache.
//
// If there is an OnEvict function, calls it for each entry.
func (c *Cache) Clear() {
	for len(c.entries) > 0 {
		c.EvictOldest()
	}
}

// EvictOldest evicts the least recently used entry from the cache.
//
// Returns the value evicted, or nil if the cache was empty.
func (c *Cache) EvictOldest() interface{} {
	if len(c.entries) == 0 {
		return nil
	}

	value := c.list.Remove(c.list.Front()).(*cell)
	delete(c.entries, value.key)
	c.cost -= value.cost
	if c.OnEvict != nil {
		c.OnEvict(value.key, value.value)
	}
	return value.value
}

// Evict evicts a specific entry from the cache.
//
// Does nothing if the entry does not exist in the cache.
//
// Calls the OnEvict function if there is one.
//
// Returns the value evicted, or nil.
func (c *Cache) Evict(key Key) interface{} {
	entry, ok := c.entries[key]
	if !ok {
		return nil
	}

	value := entry.Value.(*cell)
	delete(c.entries, value.key)
	c.list.Remove(entry)
	if c.OnEvict != nil {
		c.OnEvict(value.key, value.value)
	}
	return value.value
}
