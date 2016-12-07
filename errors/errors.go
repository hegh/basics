package errors

import (
	"bytes"
	"fmt"
	"runtime"
)

var (
	// StackDepth is the depth to which to take stack traces.
	// Functions in this library are not included at the tip of the trace.
	StackDepth int = 12
)

type detailedError struct {
	s        string
	stack    []uintptr
	cause    error
	original error
}

// Error returns the message associated with the detailedError.
func (e *detailedError) Error() string {
	return e.s
}

// Stack returns the stack trace taken at the time the error was constructed.
//
// Will not be nil, and should not be empty.
func (e *detailedError) Stack() []uintptr {
	return e.stack
}

// Original returns the error that produced it via NewTrace.
//
// A chain of NewTrace calls on successive error values will always return the
// very first value from the chain.
func (e *detailedError) Original() error {
	if e.original != nil {
		return e.original
	}
	return e
}

// NewTrace returns a copy of the errog with a new stack trace beginning skip
// frames up.
//
// skip = 0 is the caller of NewTrace.
func (e *detailedError) NewTrace(skip int) error {
	cp := new(detailedError)
	*cp = *e
	cp.stack = stackTrace(skip + 1)
	cp.original = e.Original()
	return cp
}

// Cause returns the cause of the errror. May be nil.
func (e *detailedError) Cause() error {
	return e.cause
}

// New builds a new error whose Error method will return the given string.
//
// This is a drop-in replacement for the standard library errors.New.
//
// The returned error will have a stack trace attached.
func New(text string) error {
	return &detailedError{
		s:     text,
		stack: stackTrace(1),
	}
}

// NewWithCause builds a new error whose Error method will return the given
// string.
//
// The returned error will have the given cause and a stack trace attached.
func NewWithCause(text string, cause error) error {
	return &detailedError{
		s:     text,
		stack: stackTrace(1),
		cause: cause,
	}
}

// Errorf returns an error with a formatted message, whose cause is set to the
// last error in the argument list (or nil if there are no errors in the list).
//
// This is a drop-in replacement for fmt.Errorf.
func Errorf(f string, args ...interface{}) error {
	e := &detailedError{
		s:     fmt.Sprintf(f, args...),
		stack: stackTrace(1),
	}
	for i := len(args) - 1; i >= 0; i-- {
		switch args[i].(type) {
		case error:
			e.cause = args[i].(error)
			return e
		default:
		}
	}
	return e
}

// A Stackable has an attached stack trace.
type Stackable interface {
	// Stack returns the stack trace attached to the value.
	Stack() []uintptr
}

// Stack returns the stack trace attached to the given error, or nil if it has
// none or is not a Stackable.
func Stack(err error) []uintptr {
	switch err.(type) {
	case Stackable:
		return err.(Stackable).Stack()
	}
	return nil
}

type Restackable interface {
	// NewTrace returns a copy of its receiver with a new stack trace.
	//
	// skip describes how far back to go in the call stack before beginning to
	// capture the trace. 0 is the caller of NewTrace.
	NewTrace(skip int) error
}

// NewTrace captures a new stack trace and attaches it to a copy of the given
// error.
//
// If the error was not created by this package and does not implement the
// Restackable interface, the returned error will be initialized as a new error
// with the same description as the given error and no cause.
//
// skip controls where the stack trace begins. 0 starts it at the caller of
// NewTrace.
func NewTrace(err error, skip int) error {
	switch err.(type) {
	case Restackable:
		return err.(Restackable).NewTrace(skip + 1)
	}
	return &detailedError{
		s:     err.Error(),
		stack: stackTrace(skip + 1),
	}
}

// A Causable has an attached cause.
type Causable interface {
	// Cause returns the cause attached to the value.
	Cause() error
}

// Cause returns the cause of the given error, or nil if it has none or is not
// a Causable.
func Cause(err error) error {
	switch err.(type) {
	case Causable:
		return err.(Causable).Cause()
	}
	return nil
}

// A Copy may have an original value that was copied and modified to produce it.
//
// This is used to test whether a given error may have come from a template that
// went through NewTrace or similar.
type Copy interface {
	// Original returns the template used to make the error.
	//
	// If there is no such value, it should return the error itself.
	Original() error
}

// Original returns the error that produced the given error.
//
// For example:
//   err1 := errors.New("message")
//   err2 := errors.NewTrace(err1, 0)
//   err3 := errors.NewTrace(err2, 0)
//
// In this situation:
//   err1 != err2 != err3
//   err1 == errors.Original(err2) == errors.Original(err3)
//   err2 != errors.Original(err3)
func Original(err error) error {
	switch err.(type) {
	case Copy:
		return err.(Copy).Original()
	}
	return err
}

// String formats and returns a full trace of the error and its cause chain.
//
// The result will look something like this:
//   Error message
//   pkg.Func()
//           path/file.go:123 +0x8b
//   pkg2.Func2()
//           path/file2.go:456 +0x2d4
//   Caused by: Error message 2
//   pkg3.Func3()
//           path/file3.go:789 +0x14c
//   Caused by: EOF
func String(err error) string {
	buf := bytes.NewBuffer(nil)
	first := true
	for ; err != nil; err = Cause(err) {
		if first {
			first = false
		} else {
			buf.WriteString("\nCaused by: ")
		}
		buf.WriteString(err.Error())
		stack := Stack(err)
		if stack != nil {
			frames := formatStack(stack)
			for i, frame := range frames {
				buf.WriteString("\n  ")
				if i%2 == 1 {
					buf.WriteString("  ")
				}
				buf.WriteString(frame)
			}
		}
	}
	return buf.String()
}

// FormatStack formats the given stack trace into strings that look like:
//   path/package.Function()
//   path/to/file.go:57 +0x123
//   path/other/package.OtherFunction()
//   path/to/different/file.go:83 +0x456
//
// When Go writes a panic stack trace, it indents every second line with a tab.
func formatStack(stack []uintptr) []string {
	result := make([]string, 0, len(stack))
	frames := runtime.CallersFrames(stack)
	for frame, ok := frames.Next(); ok; frame, ok = frames.Next() {
		result = append(result, fmt.Sprintf(frame.Function+"()"))
		result = append(result, fmt.Sprintf("%s:%d +0x%x", frame.File, frame.Line, frame.PC-frame.Entry))
	}
	return result
}

// StackTrace captures a stack trace.
//
// skip = 0 captures a trace starting at the caller of stackTrace.
func stackTrace(skip int) []uintptr {
	stack := make([]uintptr, StackDepth)
	stack = stack[0:runtime.Callers(skip+2, stack)]
	return stack
}
