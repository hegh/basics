package errors

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	e := New("message string")
	if m := e.Error(); m != "message string" {
		t.Errorf("Got %q want %q for error message", m, "message string")
	}

	if len(Stack(e)) == 0 {
		t.Errorf("Got 0 want some data for stack trace")
	}

	if err := Cause(e); err != nil {
		t.Errorf("Got %q want nil for error cause", err)
	}
}

func TestNewWithCause(t *testing.T) {
	ex := fmt.Errorf("the cause")
	e := NewWithCause("message string", ex)
	if m := e.Error(); m != "message string" {
		t.Errorf("Got %q want %q for error message", m, "message string")
	}

	if len(Stack(e)) == 0 {
		t.Errorf("Got 0 want some data for stack trace")
	}

	if err := Cause(e); err != ex {
		t.Errorf("Got %q want %q for error cause", err, ex)
	}
}

func TestErrorfWithoutCause(t *testing.T) {
	e := Errorf("message string: %d", 5)
	if m := e.Error(); m != "message string: 5" {
		t.Errorf("Got %q want %q for error message", m, "message string: 5")
	}

	if len(Stack(e)) == 0 {
		t.Errorf("Got 0 want some data for stack trace")
	}

	if err := Cause(e); err != nil {
		t.Errorf("Got %q want nil for error cause", err)
	}
}

func TestErrorfWithCause(t *testing.T) {
	ex := fmt.Errorf("the cause")
	e := Errorf("message string: %v", ex)
	if m := e.Error(); m != "message string: the cause" {
		t.Errorf("Got %q want %q for error message", m, "message string: the cause")
	}

	if len(Stack(e)) == 0 {
		t.Errorf("Got 0 want some data for stack trace")
	}

	if err := Cause(e); err != ex {
		t.Errorf("Got %q want %q for error cause", err, ex)
	}
}

func TestErrorfWithMultipleErrors(t *testing.T) {
	ex := fmt.Errorf("the cause")
	ex2 := fmt.Errorf("extra error")
	e := Errorf("message string (%v): %v", ex2, ex)
	if m := e.Error(); m != "message string (extra error): the cause" {
		t.Errorf("Got %q want %q for error message", m, "message string (extra error): the cause")
	}

	if len(Stack(e)) == 0 {
		t.Errorf("Got 0 want some data for stack trace")
	}

	if err := Cause(e); err != ex {
		t.Errorf("Got %q want %q for error cause", err, ex)
	}
}

// TestStackable verifies returned errors are Stackable, and that Stack()
// returns the correct result.
func TestStackable(t *testing.T) {
	err1 := New("msg")
	var trace1 []uintptr
	if s, ok := err1.(Stackable); !ok {
		t.Fatalf("value from New is not Stackable: %t", err1)
	} else {
		trace1 = s.Stack()
		if len(trace1) == 0 {
			t.Errorf("got empty trace from Stack()")
		}
	}

	trace2 := Stack(err1)
	if !reflect.DeepEqual(trace1, trace2) {
		t.Errorf("stack from err.(Stackable).Stack() %q != stack from Stack(err) %q", trace1, trace2)
	}
}

// TestRestackableOriginal verifies returned errors are Restackable, that
// NewTrace() returns an updated error, and that Original() returns the original
// error.
func TestRestackableOriginal(t *testing.T) {
	err1 := New("msg")
	var err2 error
	if s, ok := err1.(Restackable); !ok {
		t.Fatalf("value from New is not Restackable: %t", err1)
	} else {
		err2 = s.NewTrace(0)
	}

	err3 := NewTrace(err2, 0)
	trace1, trace2, trace3 := Stack(err1), Stack(err2), Stack(err3)
	if reflect.DeepEqual(trace1, trace2) {
		t.Errorf("expected unequal traces 1 and 2 %q and %q", trace1, trace2)
	}
	if reflect.DeepEqual(trace2, trace3) {
		t.Errorf("expected unequal traces 2 and 3 %q and %q", trace2, trace3)
	}
	if reflect.DeepEqual(trace1, trace3) {
		t.Errorf("expected unequal traces 1 and 3 %q and %q", trace1, trace3)
	}

	if o := Original(err1); o != err1 {
		t.Errorf("got %q want %q for Original(err1)", o, err1)
	}
	if o := Original(err2); o != err1 {
		t.Errorf("got %q want %q for Original(err2)", o, err1)
	}
	if o := Original(err3); o != err1 {
		t.Errorf("got %q want %q for Original(err3)", o, err1)
	}
}

// TestExternalErrorCause verifies Cause works correctly when the cause error is
// from a different package.
func TestExternalErrorCause(t *testing.T) {
	err1 := fmt.Errorf("msg")
	err2 := NewWithCause("msg2", err1)
	if err := Cause(err2); err != err1 {
		t.Errorf("got %q want %q for Cause(err2)", err, err1)
	}

	err3 := Errorf("msg3: %v", err1)
	if err := Cause(err3); err != err1 {
		t.Errorf("got %q want %q for Cause(err3)", err, err1)
	}
}

// TestString writes an example error to stdout for the person running the test
// to verify. It is really intended to help verify the format is readable, which
// cannot be verified by an automated test.
func TestString(t *testing.T) {
	err := io.EOF
	err = Errorf("foo failed: %v", err)
	err = Errorf("bar failed: %v", err)

	fmt.Println()
	fmt.Println("HEY YOU! TESTER! Make sure this stack trace is easy to read:")
	fmt.Println(String(err))
	fmt.Println()
}
