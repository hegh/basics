package semaphore

import (
	"testing"
	"time"
)

// We're trying to detect lockups. This is the amount of time we will sleep
// waiting for something to happen before deciding it is dead.
const timerDelay = 100 * time.Millisecond

func testSemaphore(t *testing.T, newFunc func(n int) Semaphore) {
	s := newFunc(4)

	// Try to acquire 3 at a time from 3 goroutines, without releasing.
	// The main test routine will release slots back to the semaphore to move
	// things along.

	// Sends a token on 'done' after each acquire succeeds.
	done := make(chan struct{}, 3)

	for i := 0; i < 3; i++ {
		go func() {
			s.Acquire(3)
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

func TestMutex(t *testing.T) {
	testSemaphore(t, func(n int) Semaphore { return NewMutex(n) })
}

func TestStrictMutex(t *testing.T) {
	testSemaphore(t, func(n int) Semaphore { return NewStrictMutex(n) })
}

func TestStrictMutexPanic_LargeAcquire(t *testing.T) {
	var s Semaphore = NewStrictMutex(1)

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected panic")
		}
	}()
	s.Acquire(2)
	t.Errorf("expected panic")
}

func TestStrictMutexPanic_SizeIncrease(t *testing.T) {
	var s Semaphore = NewStrictMutex(1)

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("expected panic")
		}
	}()
	s.Release(1)
	t.Errorf("expected panic")
}
