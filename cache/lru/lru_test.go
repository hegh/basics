package lru

import (
	"fmt"
	"testing"
)

func TestGetCachedEntry(t *testing.T) {
	// Verify cached entries are retrieved from the cache.
	one, two := "one", "two"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})

	// Try 1, twice
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want { // Hasn't increased.
		t.Errorf("got %v want %v calls", got, want)
	}

	// Try 2, twice
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want { // Hasn't increased.
		t.Errorf("got %v want %v calls", got, want)
	}
}

func TestErrorNotInserted(t *testing.T) {
	// Verifies that if the retriever returns an error, an entry is not inserted
	// into the cache.
	one, two := "one", "two"
	calls := 0
	var fail bool
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		if fail {
			return nil, 0, fmt.Errorf("told to fail")
		}

		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})

	// Fail inserting 1.
	fail = true
	if _, err := c.Get(1); err == nil {
		t.Errorf("expected error %v", err)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Succeed inserting 1, and make sure it required another call.
	fail = false
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want { // Hasn't increased.
		t.Errorf("got %v want %v calls", got, want)
	}
}

func TestEvictOldEntry(t *testing.T) {
	// Verify old entries get evicted.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Evict entry 1 by inserting 3.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != three {
		t.Errorf("got %v want %v", v, one)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Try to get 1 again, make sure it is re-retrieved.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != one {
		t.Errorf("got %v want %v", v, one)
	}
	if got, want := calls, 4; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
}

func TestEvictCallsEvict(t *testing.T) {
	// Verify that cache eviction calls the OnEvict function.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Evict entry 1 by inserting 3.
	evicted := false
	c.OnEvict = func(key Key, value interface{}) {
		evicted = true
		if got, want := key, 1; got != want {
			t.Errorf("got %v want %v as evicted key", got, want)
		}
		if got, want := value, one; got != want {
			t.Errorf("got %v want %v as evicted value", got, want)
		}
	}
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != three {
		t.Errorf("got %v want %v", v, one)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if !evicted {
		t.Errorf("value not evicted")
	}

	// Try to get 1 again, make sure 2 gets evicted.
	evicted = false
	c.OnEvict = func(key Key, value interface{}) {
		evicted = true
		if got, want := key, 2; got != want {
			t.Errorf("got %v want %v as evicted key", got, want)
		}
		if got, want := value, two; got != want {
			t.Errorf("got %v want %v as evicted value", got, want)
		}
	}
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != one {
		t.Errorf("got %v want %v", v, one)
	}
	if got, want := calls, 4; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if !evicted {
		t.Errorf("value not evicted")
	}
}

func TestAccessPromotesEntry(t *testing.T) {
	// Verifies that accessing an entry already in the cache promotes it so it
	// is not the next evicted.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Retrieve 1 again so it is not the least recently used.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Evict entry 2 by inserting 3.
	evicted := false
	c.OnEvict = func(key Key, value interface{}) {
		evicted = true
		if got, want := key, 2; got != want {
			t.Errorf("got %v want %v as evicted key", got, want)
		}
		if got, want := value, two; got != want {
			t.Errorf("got %v want %v as evicted value", got, want)
		}
	}
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != three {
		t.Errorf("got %v want %v", v, one)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if !evicted {
		t.Errorf("value not evicted")
	}
}

func TestClear(t *testing.T) {
	// Verify that clearing the cache actually clears it.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Clear the cache.
	evictions := 0
	c.OnEvict = func(Key, interface{}) {
		evictions++
	}
	c.Clear()
	if got, want := evictions, 2; got != want {
		t.Errorf("got %v want %v evictions after Clear", got, want)
	}

	// Make sure retrieving entries requires new lookups. Nothing new should be
	// evicted.
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// First 1, because it was already have present but should be gone.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Then 3, because it was not present, and this shouldn't cause eviction.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 4; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
}

func TestIncreaseMaxEntries(t *testing.T) {
	// Verify we can increase MaxEntries on the fly.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(1, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// 2 should evict 1.
	evicted := false
	c.OnEvict = func(key Key, value interface{}) {
		evicted = true
		if got, want := key, 1; got != want {
			t.Errorf("got %v want %v as evicted key", got, want)
		}
		if got, want := value, one; got != want {
			t.Errorf("got %v want %v as evicted value", got, want)
		}
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if !evicted {
		t.Errorf("expected an eviction")
	}

	// Increase the capacity, and make sure 1 gets re-read, and nothing gets
	// evicted.
	c.MaxCost = 3
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 4; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
}

func TestDecreaseMaxEntries(t *testing.T) {
	// Verify we can decrease MaxEntries on the fly.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, 1, nil
		case 2:
			return two, 1, nil
		case 3:
			return three, 1, nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}

	// Reduce the max size.
	c.MaxCost = 1

	// The next new Get should evict 1 and 2.
	evictions := 0
	c.OnEvict = func(key Key, value interface{}) {
		evictions++
		if got, want1, want2 := key, 1, 2; got != want1 && got != want2 {
			t.Errorf("got %v want %v or %v as evicted key", got, want1, want2)
		}
		if got, want1, want2 := value, one, two; got != want1 && got != want2 {
			t.Errorf("got %v want %v or %v as evicted value", got, want1, want2)
		}
	}

	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 2; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}
}

func TestCostBasedEviction(t *testing.T) {
	// Verify values are evicted based on cost.
	one, two, three := "one", "two", "three"

	// 6 is big enough to hold one and two, or three, but not three and any other
	// value.
	calls := 0
	c := New(6, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, Cost(len(one)), nil
		case 2:
			return two, Cost(len(two)), nil
		case 3:
			return three, Cost(len(three)), nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})

	evictions := 0
	c.OnEvict = func(Key, interface{}) { evictions++ }

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 0; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 0; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// Inserting 3 should evict both 1 and 2.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 2; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// Getting 1 should evict 3.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 4; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 3; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// Getting 2 should not evict anything.
	c.OnEvict = func(key Key, value interface{}) {
		// Use panic because t.Errorf doesn't tell us where it happened.
		panic(fmt.Errorf("unexpected eviction of key %v value %v", key, value))
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 5; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 3; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}
}

func TestJumboEntry(t *testing.T) {
	// Verify the cache will hold onto a single jumbo entry.
	one, two, three := "one", "two", "three"

	// 4 is big enough to hold one or two, but not three unless it's the
	// singular jumbo entry.
	calls := 0
	c := New(4, func(key Key) (interface{}, Cost, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, Cost(len(one)), nil
		case 2:
			return two, Cost(len(two)), nil
		case 3:
			return three, Cost(len(three)), nil
		default:
			return nil, 0, fmt.Errorf("bad key %v", key)
		}
	})

	evictions := 0
	c.OnEvict = func(Key, interface{}) { evictions++ }

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 1; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 0; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// Inserting 3 should evict 1.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 1; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// 3 should remain in the cache, so no change by getting it again.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 2; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 1; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}

	// Getting 1 should evict 3.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := calls, 3; got != want {
		t.Errorf("got %v want %v calls", got, want)
	}
	if got, want := evictions, 2; got != want {
		t.Errorf("got %v want %v evictions", got, want)
	}
}
