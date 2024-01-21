// Deprecated. Prefer using the standard library `errors` package instead.
//
// Package errors provides additional functionality beyond the Go standard
// library, focusing on including stack trace data and causes, like Java's
// exceptions (after Java 5), even through formats.
//
// Provides drop-in replacements for errors.New and fmt.Errorf that capture
// and retain more detailed error information.
//
// To format an error string including a stack trace and causes when they are
// available (works fine on other errors too):
//
//	errors.String(err)
//
// Example when returning an error:
//
//	return errors.New("message")
//
// Example for reuse of a common package-level error:
//
//	var pkgError = errors.New("message")
//	return errors.NewTrace(pkgError, 0)
//
// Example returning a formatted error that includes the original error as its
// cause:
//
//	return errors.Errorf("foo failed: %v", err)
//
// Example testing whether the root cause of an error is an io.EOF:
//
//	func isEOF(err error) bool {
//	  for err != nil {
//	    if err == io.EOF { return true }
//	    err = errors.Cause(err)
//	  }
//	  return false
//	}
package errors
