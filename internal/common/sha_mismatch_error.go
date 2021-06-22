package common

import "fmt"

type shaMismatch interface {
	ShaMismatch() bool
}

// ShaMismatchError error is raised when the downloaded kubectl's SHA
// doesn't match the recorded SHA
type ShaMismatchError struct {
	URL         string
	ShaExpected string
	ShaActual   string
}

// Error returns a human description of the error
func (e *ShaMismatchError) Error() string {
	return fmt.Sprintf("SHA mismatch for URL %s: expected '%s', got '%s'", e.URL, e.ShaExpected, e.ShaActual)
}

// ShaMismatch returns true if the error is a ShaMismatchError instance
func (e *ShaMismatchError) ShaMismatch() bool {
	return true
}

// IsShaMismatch returns true when the given error is of type
// ShaMismatchError
func IsShaMismatch(err error) bool {
	t, ok := err.(shaMismatch)
	return ok && t.ShaMismatch()
}
