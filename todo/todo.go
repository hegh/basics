// Package todo allows unimplemented bits of code to provide helpful errors when
// they get called.
package todo

import (
	"fmt"
	"runtime"

	"github.com/hegh/basics/errors"
)

func Error() error {
	return makeError()
}

func Panic() {
	panic(makeError())
}

func makeError() error {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return errors.New("TODO: implement unknown func")
	}

	f := runtime.FuncForPC(pc)
	if f == nil {
		return fmt.Errorf("TODO: implement unknown func at %s:%d", file, line)
	}

	return fmt.Errorf("TODO: implement func %s at %s:%d", f.Name(), file, line)
}
