package errors

import (
	"fmt"
	"runtime"
	"strconv"
)

var (
	// StackDepth is the depth to which to take stack traces.
	// Functions in this library are not included at the tip of the trace.
	StackDepth int = 12
)

type detailedError struct {
	s     string
	stack []uintptr
	cause error
}

// New builds a new error whose Error method will return the given string.
// The returned error will have a stack trace attached.
func New(text string) error {
	stack := make([]uintptr, StackDepth)
	stack = stack[0:runtime.Callers(2, stack)]
	return &detailedError{
		s:     text,
		stack: stack,
	}
}

// NewWithCause builds a new error whose Error method will return the given
// string.
// The returned error will have the given cause and a stack trace attached.
func NewWithCause(text string, cause error) error {
	stack := make([]uintptr, StackDepth)
	stack = stack[0:runtime.Callers(2, stack)]
	return &detailedError{
		s:     text,
		stack: stack,
		cause: cause,
	}
}

// Print builds a new error whose Error method will return the concatenation of
// string values of all of the arguments, with spaces between non-string args.
// If any errors are present in the arguments, the returned error's cause will
// be set to the last one.
func Print(args ...interface{}) error {
	stack := make([]uintptr, StackDepth)
	stack = stack[0:runtime.Callers(2, stack)]
	e := &detailedError{
		s:     fmt.Sprint(args...),
		stack: stack,
	}
	for i := len(args) - 1; i >= 0; i-- {
		switch args[i].(type) {
		case error:
			e.cause = args[i].(error)
			return e
		}
	}
	return e
}

// Printf returns an error with a formatted message, whose cause is set to the
// last error in the argument list (or nil if there are no errors in the list).
func Printf(f string, args ...interface{}) error {
	stack := make([]uintptr, StackDepth)
	stack = stack[0:runtime.Callers(2, stack)]
	e := &detailedError{
		s:     fmt.Sprintf(f, args...),
		stack: stack,
	}
	for i := len(args) - 1; i >= 0; i-- {
		switch args[i].(type) {
		case error:
			e.cause = args[i].(error)
			return e
		}
	}
	return e
}

func (e *detailedError) Error() string {
	return e.s
}

// Stack returns the stack trace taken at the time the error was constructed.
// Will not be nil, and should not be empty.
func (e *detailedError) Stack() []uintptr {
	return e.stack
}

// Cause returns the cause of the errror. May be nil.
func (e *detailedError) Cause() error {
	return e.cause
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

// FormatStack formats the given stack trace into strings of filename:line#.
// Any PC values in the stack that cannot be resolved will be formatted as
// "?????:??".
func FormatStack(stack []uintptr) []string {
	result := make([]string, 0, len(stack))
	for _, pc := range stack {
		f := runtime.FuncForPC(pc)
		if f == nil {
			result = append(result, "?????:??")
			continue
		}
		file, line := f.FileLine(pc)
		result = append(result, file+":"+strconv.Itoa(line))
	}
	return result
}
