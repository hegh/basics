package writecounter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type limitWriter struct {
	writeTo   io.Writer
	remaining int
}

func (w *limitWriter) Write(p []byte) (n int, err error) {
	if w.remaining > len(p) {
		n, err = w.writeTo.Write(p)
		w.remaining -= n
		return
	}
	n, err = w.writeTo.Write(p[:w.remaining])
	w.remaining -= n
	if err != nil {
		return
	}
	err = fmt.Errorf("hit write limit")
	return
}

func TestWriteCount(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	// Start at 0.
	if got, want := w.Count(), int64(0); got != want {
		t.Errorf("got %d want %d from Count before Write", got, want)
	}

	// Write 5 bytes, make sure they get counted correctly.
	if n, err := w.Write([]byte{1, 2, 3, 4, 5}); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got, want := n, 5; got != want {
		t.Errorf("got %d want %d bytes written from Write", got, want)
	} else if got, want := buf.Len(), 5; got != want {
		t.Errorf("got %d want %d bytes in output after Write", got, want)
	}
	if got, want := w.Count(), int64(5); got != want {
		t.Errorf("got %d want %d from Count after Write", got, want)
	}

	// Write 3 more bytes, make sure they also get counted correctly.
	if n, err := w.Write([]byte{6, 7, 8}); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if got, want := n, 3; got != want {
		t.Errorf("got %d want %d bytes written from Write", got, want)
	}
	if got, want := buf.Len(), 8; got != want {
		t.Errorf("got %d want %d bytes in output after Write", got, want)
	}
	if got, want := w.Count(), int64(8); got != want {
		t.Errorf("got %d want %d from Count after Write", got, want)
	}

	// Make sure everything wrote through clearly.
	if got, want := buf.Bytes(), ([]byte{1, 2, 3, 4, 5, 6, 7, 8}); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%x want\n%x as written content", got, want)
	}
}

func TestWriteWithError(t *testing.T) {
	// Verify if there's a partial write with an error, the Count is incremented
	// correctly.
	var buf bytes.Buffer
	lw := &limitWriter{&buf, 5}
	w := New(lw)

	// Write 8 bytes, and expect only 5 to make it through and be counted.
	if n, err := w.Write([]byte{1, 2, 3, 4, 5, 6, 7, 8}); err == nil {
		t.Errorf("expected error")
	} else if got, want := n, 5; got != want {
		t.Errorf("got %d want %d bytes written", got, want)
	}
	if got, want := buf.Len(), 5; got != want {
		t.Errorf("got %d want %d bytes in output after Write", got, want)
	}
	if got, want := w.Count(), int64(5); got != want {
		t.Errorf("got %d want %d from Count after Write", got, want)
	}

	// Make sure everything got through clearly.
	if got, want := buf.Bytes(), ([]byte{1, 2, 3, 4, 5}); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%x want\n%x as written content", got, want)
	}
}

func TestWriteValue_DefaultBigEndian(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)

	// Verify the default endian-ness is big
	if n, err := w.WriteValue(uint64(0x0123456789abcdef)); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := n, 8; got != want {
		t.Errorf("got %d want %d bytes written", got, want)
	}
	if got, want := w.Count(), int64(8); got != want {
		t.Errorf("got %d want %d from Count after 8-byte write", got, want)
	}
	if got, want := buf.Bytes(), ([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%x want\n%x from write of uint64", got, want)
	}
}

func TestWriteValue_LittleEndian(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf)
	w.ByteOrder = binary.LittleEndian

	// Verify little-endian works too.
	if n, err := w.WriteValue(uint32(0x01234567)); err != nil {
		t.Errorf("unexpected error %v", err)
	} else if got, want := n, 4; got != want {
		t.Errorf("got %d want %d bytes written", got, want)
	}
	if got, want := w.Count(), int64(4); got != want {
		t.Errorf("got %d want %d from Count after 4-byte write", got, want)
	}
	if got, want := buf.Bytes(), ([]byte{0x67, 0x45, 0x23, 0x01}); !reflect.DeepEqual(got, want) {
		t.Errorf("got\n%x want\n%x from write of uint32", got, want)
	}
}
