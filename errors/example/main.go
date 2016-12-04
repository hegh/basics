///bin/true; exec /usr/bin/env go run "$0" "$@"
// An example program to demonstrate the errors package.
package main

import (
	"fmt"
	"io"

	"github.com/hegh/basics/errors"
)

func ReadData() error {
	return io.EOF
}

func Foo() error {
	if err := ReadData(); err != nil {
		return errors.Errorf("failure in ReadData: %v", err)
	}
	return nil
}

func Bar() error {
	if err := Foo(); err != nil {
		return errors.Errorf("failure in Foo: %v", err)
	}
	return nil
}

func main() {
	err := ReadData()
	fmt.Printf("ReadData returned an error: %v\n", errors.String(err))
	fmt.Println()
	err = Bar()
	fmt.Printf("Bar returned an error: %v\n", errors.String(err))
}
