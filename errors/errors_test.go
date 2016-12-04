package errors

import (
	"fmt"
	"io"
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
}
