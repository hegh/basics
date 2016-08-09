package errors

import (
	"fmt"
	"runtime"
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

func TestPrintWithoutCause(t *testing.T) {
	e := Print("message string: ", 5)
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

func TestPrintWithCause(t *testing.T) {
	ex := fmt.Errorf("the cause")
	e := Print("message string: ", ex)
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

func TestPrintWithMultipleErrors(t *testing.T) {
	ex := fmt.Errorf("the cause")
	ex2 := fmt.Errorf("extra error")
	e := Print("message string (", ex2, "): ", ex)
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

func TestPrintfWithoutCause(t *testing.T) {
	e := Printf("message string: %d", 5)
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

func TestPrintfWithCause(t *testing.T) {
	ex := fmt.Errorf("the cause")
	e := Printf("message string: %v", ex)
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

func TestPrintfWithMultipleErrors(t *testing.T) {
	ex := fmt.Errorf("the cause")
	ex2 := fmt.Errorf("extra error")
	e := Printf("message string (%v): %v", ex2, ex)
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

func TestFormatStack(t *testing.T) {
	// Keep these two lines together, in this order, to avoid breaking.
	_, file, line, ok := runtime.Caller(0)
	err := New("message")

	if !ok {
		t.Fatalf("Could not determine location of caller")
	}

	trace := FormatStack(Stack(err))
	want := fmt.Sprintf("%s:%d", file, line+1)
	if trace[0] != want {
		t.Errorf("Got %q want %q for first element of formatted stack trace. Full trace: %v", trace[0], want, trace)
	}
}
