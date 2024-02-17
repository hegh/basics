// Package semaphore provides some implementations of a semaphore (basically a
// mutex that can be acquired N times before blocking).
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
	// Panics if `n` is zero or negative.
	//
	// Check the specific implementation for behavior when trying to acquire more
	// slots than the semaphore was created with; this is unspecified by the
	// interface.
	Acquire(n int)

	// Release releases `n` semaphore slots, so they may be acquired by others.
	//
	// Panics if `n` is zero or negative.
	//
	// Check the specific implementation for behavior when trying to release more
	// slots than the semaphore was created with; this is unspecified by the
	// interface.
	Release(n int)
}

// Mutex implements Semaphore using a mutex and condition variable.
//
// Takes constant time to acquire or release N slots.
//
// Releasing slots you have not acquired will increase the size of the
// semaphore.
// Acquiring slots and never releasing them will decrease the size of the
// semaphore.
// Acquiring more slots than the semaphore can provide will block forever.
type Mutex struct {
	lock  sync.Mutex
	cond  *sync.Cond
	slots int
}

func NewMutex(size int) *Mutex {
	s := &Mutex{slots: size}
	s.cond = sync.NewCond(&s.lock)
	return s
}

func (s *Mutex) Acquire(n int) {
	if n <= 0 {
		panic(fmt.Errorf("cannot acquire %d <= 0 slots", n))
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	for s.slots < n {
		s.cond.Wait()
	}
	s.slots -= n
}

func (s *Mutex) Release(n int) {
	s.release(n)
}
func (s *Mutex) release(n int) int { // Used by StrictMutex.
	if n <= 0 {
		panic(fmt.Errorf("cannot release %d <= 0 slots", n))
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.slots += n
	s.cond.Broadcast()
	return s.slots
}

// StrictMutex is a Mutex semaphore that disallows size changes.
//
// Panics if Release would increase the size of the semaphore beyond its created
// size.
// Panics if Acquire is called with `n` greater than the created size.
type StrictMutex struct {
	s    *Mutex
	base int
}

func NewStrictMutex(size int) *StrictMutex {
	return &StrictMutex{
		s:    NewMutex(size),
		base: size,
	}
}

func (s *StrictMutex) Acquire(n int) {
	if n > s.base {
		panic(fmt.Errorf("cannot acquire %d > base size %d slots", n, s.base))
	}
	s.s.Acquire(n)
}

func (s *StrictMutex) Release(n int) {
	if x := s.s.release(n); x > s.base {
		panic(fmt.Errorf("released %d slots, increasing size to %d > base size %d", n, x, s.base))
	}
}
