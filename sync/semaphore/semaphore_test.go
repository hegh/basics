package semaphore

import (
	"testing"
	"time"
)

// We're trying to detect lockups. This is the amount of time we will sleep
// waiting for something to happen before deciding it is dead.
const timerDelay = 100 * time.Millisecond

func testAcquireRelease(t *testing.T, newFunc func(n int) Semaphore) {
	s := newFunc(4)

	// Try to acquire 3 at a time from 3 goroutines, without releasing.
	// The main test routine will release slots back to the semaphore to move
	// things along.

	// Sends a token on 'done' after each acquire succeeds.
	done := make(chan struct{}, 3)

	for i := 0; i < 3; i++ {
		go func() {
			if n := s.Acquire(3); n != 3 {
				t.Errorf("got %d want 3 from Acquire", n)
			}
			done <- struct{}{}
		}()
	}

	// First should succeed immediately.
	select {
	case <-done:
		// Good.
	case <-time.After(timerDelay):
		t.Fatalf("first acquisition failed")
	}

	// Second should still be blocked.
	select {
	case <-done:
		t.Fatalf("second acquisition fired early")
	case <-time.After(timerDelay):
		// Good.
	}

	// Add 1 (up to 2), which should not be enough to move forward.
	s.Release(1)
	select {
	case <-done:
		t.Fatalf("second acquisition fired early")
	case <-time.After(timerDelay):
		// Good.
	}

	// Add 1 (up to 3), which should allow movement.
	s.Release(1)
	select {
	case <-done:
		// Good.
	case <-time.After(timerDelay):
		t.Fatalf("second acquisition failed")
	}

	// Third should still be blocked.
	select {
	case <-done:
		t.Fatalf("third acquisition fired early")
	case <-time.After(timerDelay):
		// Good.
	}

	// Add 4 (up to 4), which should allow movement.
	s.Release(4)
	select {
	case <-done:
		// Good.
	case <-time.After(timerDelay):
		t.Fatalf("third acquisition failed")
	}
}

func testClose(t *testing.T, newFunc func(n int) Semaphore) {
	s := newFunc(3)

	// Grab one of the slots and never return it, so the other acquires block
	// without trying to acquire more than the semaphore size.
	if n := s.Acquire(1); n != 1 {
		t.Fatalf("got %d want 1 from Acquire before Close", n)
	}

	// Try to acquire 3 at a time from 3 goroutines, without releasing.
	// All of these should block, until we close the semaphore.

	// Sends a token on 'done' after each acquire succeeds.
	done := make(chan struct{}, 3)

	for i := 0; i < 3; i++ {
		go func() {
			n := s.Acquire(3)
			if n != 0 {
				t.Errorf("got %d want 0 from Acquire after Close", n)
			}
			done <- struct{}{}
		}()
	}

	// Make sure nothing has been acquired.
	select {
	case <-done:
		t.Fatalf("acquisition fired early")
	case <-time.After(timerDelay):
		// Good.
	}

	// Close the semaphore and make sure everyone immediately returns.
	if err := s.Close(); err != nil {
		t.Fatalf("unexpected error from Close: %v", err)
	}

	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Good.
		case <-time.After(timerDelay):
			t.Fatalf("all acquisitions should have succeeded")
		}
	}

	// Make sure a new acquisition immediately succeeds.
	go func() {
		n := s.Acquire(1)
		if n != 0 {
			t.Errorf("got %d want 0 from Acquire after Close", n)
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
		// Good.
	case <-time.After(timerDelay):
		t.Errorf("new acquisition should have succeeded")
	}
}

func TestRegularAcquireRelease(t *testing.T) {
	testAcquireRelease(t, func(n int) Semaphore { return New(n) })
}
func TestRegularClose(t *testing.T) {
	testClose(t, func(n int) Semaphore { return New(n) })
}

func TestStrictAcquireRelease(t *testing.T) {
	testAcquireRelease(t, func(n int) Semaphore { return NewStrict(n) })
}
func TestStrictClose(t *testing.T) {
	testClose(t, func(n int) Semaphore { return NewStrict(n) })
}

func TestStrictPanic_LargeAcquire(t *testing.T) {
	var s Semaphore = NewStrict(1)

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected panic")
		}
	}()
	s.Acquire(2)
	t.Errorf("expected panic")
}

func TestStrictPanic_SizeIncrease(t *testing.T) {
	var s Semaphore = NewStrict(1)

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected panic")
		}
	}()
	s.Release(1)
	t.Errorf("expected panic")
}
