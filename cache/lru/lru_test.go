package lru

import (
	"fmt"
	"testing"
)

func TestGetCachedEntry(t *testing.T) {
	// Verify cached entries are retrieved from the cache.
	one, two := "one", "two"
	calls := 0
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})

	// Try 1, twice
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}

	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 { // Hasn't increased.
		t.Errorf("got %v want 1 calls", calls)
	}

	// Try 2, twice
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
	}

	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 { // Hasn't increased.
		t.Errorf("got %v want 2 calls", calls)
	}
}

func TestErrorNotInserted(t *testing.T) {
	// Verifies that if the retriever returns an error, an entry is not inserted
	// into the cache.
	one, two := "one", "two"
	calls := 0
	var fail bool
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		if fail {
			return nil, fmt.Errorf("told to fail")
		}

		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})

	// Fail inserting 1.
	fail = true
	if _, err := c.Get(1); err == nil {
		t.Errorf("expected error %v", err)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}

	// Succeed inserting 1, and make sure it required another call.
	fail = false
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 { // Hasn't increased.
		t.Errorf("got %v want 2 calls", calls)
	}
}

func TestEvictOldEntry(t *testing.T) {
	// Verify old entries get evicted.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
	}

	// Evict entry 1 by inserting 3.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != three {
		t.Errorf("got %v want %v", v, one)
	}
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
	}

	// Try to get 1 again, make sure it is re-retrieved.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if v != one {
		t.Errorf("got %v want %v", v, one)
	}
	if calls != 4 {
		t.Errorf("got %v want 4 calls", calls)
	}
}

func TestEvictCallsEvict(t *testing.T) {
	// Verify that cache eviction calls the OnEvict function.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
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
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
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
	if calls != 4 {
		t.Errorf("got %v want 4 calls", calls)
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
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
	}

	// Retrieve 1 again so it is not the least recently used.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
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
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
	}
	if !evicted {
		t.Errorf("value not evicted")
	}
}

func TestClear(t *testing.T) {
	// Verify that clearing the cache actually clears it.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
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
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// First 1, because it was already have present but should be gone.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
	}

	// Then 3, because it was not present, and this shouldn't cause eviction.
	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 4 {
		t.Errorf("got %v want 4 calls", calls)
	}
}

func TestIncreaseMaxEntries(t *testing.T) {
	// Verify we can increase MaxEntries on the fly.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(1, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
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
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
	}

	// Increase the capacity, and make sure 1 gets re-read, and nothing gets
	// evicted.
	c.MaxEntries = 3
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
	}

	if v, err := c.Get(3); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, three; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 4 {
		t.Errorf("got %v want 4 calls", calls)
	}
}

func TestDecreaseMaxEntries(t *testing.T) {
	// Verify we can decrease MaxEntries on the fly.
	one, two, three := "one", "two", "three"
	calls := 0
	c := New(2, func(key Key) (interface{}, error) {
		calls++
		switch key.(int) {
		case 1:
			return one, nil
		case 2:
			return two, nil
		case 3:
			return three, nil
		default:
			return nil, fmt.Errorf("bad key %v", key)
		}
	})
	c.OnEvict = func(key Key, value interface{}) {
		t.Errorf("unexpected eviction of key %v value %v", key, value)
	}

	// Populate the cache.
	if v, err := c.Get(1); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, one; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("got %v want 1 calls", calls)
	}
	if v, err := c.Get(2); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := v, two; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if calls != 2 {
		t.Errorf("got %v want 2 calls", calls)
	}

	// Reduce the max size.
	c.MaxEntries = 1

	// The next new Get should evict 1 and 2.
	evicted := 0
	c.OnEvict = func(key Key, value interface{}) {
		evicted++
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
	if calls != 3 {
		t.Errorf("got %v want 3 calls", calls)
	}
	if got, want := evicted, 2; got != want {
		t.Errorf("got %d want %d evictions", got, want)
	}
}
