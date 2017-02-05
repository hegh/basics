package todo

import (
	"fmt"
	"runtime"
	"testing"
)

func TestError(t *testing.T) {
	// XXX: These two statements need to be next to each other.
	err := Error()
	pc, file, line, ok := runtime.Caller(0)
	line-- // So it matches the Error() call.

	if err == nil {
		t.Fatal("got nil, want error from Error()")
	}
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	f := runtime.FuncForPC(pc)
	if f == nil {
		t.Fatal("runtime.FuncForPC() failed")
	}

	want := fmt.Sprintf("TODO: implement func %s at %s:%d", f.Name(), file, line)
	if err.Error() != want {
		t.Errorf("got\n%v, want\n%v from Error()", err, want)
	}
}

func TestPanic(t *testing.T) {
	var pc uintptr
	var file string
	var line int
	var callerOk bool
	defer func() {
		x := recover()
		if x == nil {
			t.Fatal("got nil, want error from deferred recover after Panic()")
		}
		err, ok := x.(error)
		if !ok {
			t.Fatalf("got %T, want error as type from recover after Panic()", x)
		}
		if !callerOk {
			t.Fatal("runtime.Caller() failed")
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			t.Fatal("runtime.FuncForPC() failed")
		}

		want := fmt.Sprintf("TODO: implement func %s at %s:%d", f.Name(), file, line)
		if err.Error() != want {
			t.Errorf("got\n%v, want\n%v from recover after Panic()", err, want)
		}
	}()

	// XXX: The line numbers here are counted.
	pc, file, line, callerOk = runtime.Caller(0)
	line += 2 // To match the Panic() call.
	Panic()

	t.Error("panic() did not panic")
}
