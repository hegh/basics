package refcount

import (
	"fmt"
	"testing"
)

type testObject struct {
	opens, closes int // Count number of calls to open & close

	// Returned from the next open/close call.
	err error
}

func (o *testObject) open() error {
	o.opens++
	return o.err
}

func (o *testObject) close() error {
	o.closes++
	return o.err
}

func TestFirstIncrementOpens(t *testing.T) {
	o := &testObject{}
	rc := New(o.open, o.close)

	if o.opens != 0 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after New", o.opens, o.closes, 0, 0)
	}

	_, err := rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 1; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after second Increment", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 2; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestLastDecrementCloses(t *testing.T) {
	o := &testObject{}
	rc := New(o.open, o.close)

	if o.opens != 0 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after New", o.opens, o.closes, 0, 0)
	}

	closer, err := rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 1; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	err = closer.Close()
	if o.opens != 1 || o.closes != 1 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Close", o.opens, o.closes, 1, 1)
	}
	if got, want := rc.Instances(), 0; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestReopen(t *testing.T) {
	o := &testObject{}
	rc := New(o.open, o.close)

	if o.opens != 0 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after New", o.opens, o.closes, 0, 0)
	}

	closer, err := rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 1; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	err = closer.Close()
	if o.opens != 1 || o.closes != 1 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Close", o.opens, o.closes, 1, 1)
	}
	if got, want := rc.Instances(), 0; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, err = rc.Increment()
	if o.opens != 2 || o.closes != 1 {
		t.Errorf("got %d %d want %d %d for opens/closes after second Increment", o.opens, o.closes, 2, 1)
	}
	if got, want := rc.Instances(), 1; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
}

func TestOpenError(t *testing.T) {
	o := &testObject{}
	rc := New(o.open, o.close)

	if o.opens != 0 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after New", o.opens, o.closes, 0, 0)
	}

	o.err = fmt.Errorf("error")
	_, err := rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment with error", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 0; got != want {
		// Open failed, so instances should remain at 0.
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != o.err {
		t.Errorf("expected error, got nil")
	}
}

func TestCloseError(t *testing.T) {
	o := &testObject{}
	rc := New(o.open, o.close)

	if o.opens != 0 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after New", o.opens, o.closes, 0, 0)
	}

	closer, err := rc.Increment()
	if o.opens != 1 || o.closes != 0 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment with error", o.opens, o.closes, 1, 0)
	}
	if got, want := rc.Instances(), 1; got != want {
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	o.err = fmt.Errorf("error")
	err = closer.Close()
	if o.opens != 1 || o.closes != 1 {
		t.Errorf("got %d %d want %d %d for opens/closes after first Increment with error", o.opens, o.closes, 1, 1)
	}
	if got, want := rc.Instances(), 1; got != want {
		// Close failed, so instances should remain at 1.
		t.Errorf("got %d want %d from Instances", got, want)
	}
	if err != o.err {
		t.Errorf("expected error, got nil")
	}
}
