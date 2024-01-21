// Package writecounter provides a simple io.Writer that counts the bytes it has
// written so far.
//
// It also has provides a convenience method for writing fixed-size data.
package writecounter

import (
	"encoding/binary"
	"io"
)

// Writer is the type that implements the write counter.
type Writer struct {
	// ByteOrder determines the byte order to use for calls to WriteValue.
	// It defaults to BigEndian.
	ByteOrder binary.ByteOrder

	w io.Writer
	n int64
}

// New returns a new Writer that will write to the given writer and count all of
// the bytes successfully written.
//
// Defaults to BigEndian byte order.
func New(w io.Writer) *Writer {
	return &Writer{
		ByteOrder: binary.BigEndian,
		w:         w,
	}
}

// Count returns the number of bytes successfully written to the underlying
// writer.
func (w *Writer) Count() int64 { return w.n }

// Write writes the data in `p` to the underlying writer, counting the number
// of bytes actually written.
func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.n += int64(n)
	return
}

// WriteValue is a convenience wrapper around `binary.Write` using the
// endianness of the Writer.
func (w *Writer) WriteValue(value interface{}) (n int, err error) {
	on := w.n
	err = binary.Write(w, w.ByteOrder, value)
	nn := w.n
	n = int(nn - on)
	return
}
