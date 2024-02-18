// Package semaphore provides some implementations of a semaphore (basically a
// mutex that can be acquired N times before blocking).
//
// In addition to the usual `Acquire` and `Release`, these semaphores can be
// `Close`d to release all waiting goroutines. This is particularly useful when
// the semaphore is being used for communication across goroutines, where one
// acquires some slots and another releases them, making `defer` unreliable.
//
// Definitions:
//   - Slot - A slot is the basic unit guarded by the semaphore. You acquire and
//     release slots using the semaphore.
//   - Size - A semaphore's size is the number of slots provided by the semaphore
//     when no slots are outstanding (waiting to be released).
//   - Acquire/Release - The basic operations on a semaphore. Acquire blocks
//     until it is able to claim the requested number of slots, and Release
//     releases slots back to the semaphore so they can be acquired.
package semaphore

import (
	"fmt"
	"sync"
)

// TODO: Is there a way to write a semaphore based on an atomic integer? Got
// stuck trying to decrement without going below 0, and how to deal with
// blocking without needing to use locks.

// TODO: Is there a way to write a semaphore based on a channel? Got stuck on
// two versions:
//  * N-buffered channel of structs: Acquisition of X slots is not atomic, so
//    even if there are X available, they may get split across Y different
//    acquirers
//  * 1-buffered channel of int (a locked int): No way to block if there aren't
//    any tokens at the moment

// Semaphore is the interface provided by the semaphore implementations in this
// package.
type Semaphore interface {
	// Acquire blocks until it has acquired `n` semaphore slots, then returns.
	//
	// Returns the number of slots acquired, probably `n` or `0` (from a closed
	// Semaphore).
	//
	// Panics if `n` is zero or negative.
	//
	// Check the specific implementation for behavior when trying to acquire more
	// slots than the semaphore was created with; this is unspecified by the
	// interface.
	Acquire(n int) int

	// Release releases `n` semaphore slots, so they may be acquired by others.
	//
	// Panics if `n` is zero or negative.
	//
	// Check the specific implementation for behavior when trying to release more
	// slots than the semaphore was created with; this is unspecified by the
	// interface.
	Release(n int)

	// Close destroys the semaphore, releasing all waiting goroutines. Calls to
	// `Acquire` will return 0 from this point on.
	//
	// Probably never returns an error, but check the specific implementation for
	// details.
	Close() error
}

// Basic implements Semaphore using a mutex and condition variable.
// This is the type of sempahore returned from `New`.
//
// Takes constant time to acquire or release N slots.
//
// Releasing slots you have not acquired will increase the size of the
// semaphore.
// Acquiring slots and never releasing them will decrease the size of the
// semaphore.
// Acquiring more slots than the semaphore can provide will block forever.
type Basic struct {
	lock   sync.Mutex
	cond   *sync.Cond
	slots  int
	closed bool
}

func New(size int) *Basic {
	s := &Basic{slots: size}
	s.cond = sync.NewCond(&s.lock)
	return s
}

// Acquire acquires `n` slots from the semaphore, blocking until enough are
// available.
//
// Returns the number of slots acquired, which will be `0` or `n` depending on
// whether the semaphore has been closed.
//
// Panics if `n <= 0`.
func (s *Basic) Acquire(n int) int {
	if n <= 0 {
		panic(fmt.Errorf("cannot acquire %d <= 0 slots", n))
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return 0
	}

	for s.slots < n && !s.closed {
		s.cond.Wait()
	}
	if s.closed {
		return 0
	}
	s.slots -= n
	return n
}

// Release releases `n` slots back to the semaphore.
//
// Panics if `n <= 0`.
func (s *Basic) Release(n int) {
	s.release(n)
}
func (s *Basic) release(n int) int { // Used by Strict.
	if n <= 0 {
		panic(fmt.Errorf("cannot release %d <= 0 slots", n))
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.closed {
		return 0
	}

	s.slots += n
	s.cond.Broadcast()
	return s.slots
}

// Close destroys the semaphore, releasing all waiting goroutines. Always
// returns nil.
func (s *Basic) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.closed {
		s.closed = true
		s.cond.Broadcast()
	}
	return nil
}

// Strict is a Basic semaphore that disallows size changes.
//
// Panics if Release would increase the size of the semaphore beyond its created
// size.
// Panics if Acquire is called with `n` greater than the created size.
type Strict struct {
	s    *Basic
	base int
}

func NewStrict(size int) *Strict {
	return &Strict{
		s:    New(size),
		base: size,
	}
}

// Acquire acquires `n` slots from the semaphore, blocking until enough are
// available.
//
// Returns the number of slots acquired, which will be `0` or `n` depending on
// whether the semaphore has been closed.
//
// Panics if `n` is greater than the initial size of the semaphore.
// Panics if `n <= 0`.
func (s *Strict) Acquire(n int) int {
	if n > s.base {
		panic(fmt.Errorf("cannot acquire %d > base size %d slots", n, s.base))
	}
	return s.s.Acquire(n)
}

// Release releases `n` slots back to the semaphore.
//
// Panics if the release increases the semaphore's size beyond its initial size.
// Panics if `n <= 0`.
func (s *Strict) Release(n int) {
	if x := s.s.release(n); x > s.base {
		panic(fmt.Errorf("released %d slots, increasing size to %d > base size %d", n, x, s.base))
	}
}

// Close destroys the semaphore, releasing all waiting goroutines. Always
// returns nil.
func (s *Strict) Close() error {
	return s.s.Close()
}
