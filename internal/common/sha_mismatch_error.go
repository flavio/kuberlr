package common

import (
	"errors"
	"fmt"
)

// ShaMismatchError error is raised when the downloaded kubectl's SHA
// doesn't match the recorded SHA.
type ShaMismatchError struct {
	URL         string
	ShaExpected string
	ShaActual   string
}

// Error returns a human description of the error.
func (e *ShaMismatchError) Error() string {
	return fmt.Sprintf("SHA mismatch for URL %s: expected '%s', got '%s'", e.URL, e.ShaExpected, e.ShaActual)
}

// IsShaMismatch returns true when the given error is of type
// ShaMismatchError.
func IsShaMismatch(err error) bool {
	var shaMismatchErr *ShaMismatchError

	return errors.As(err, &shaMismatchErr)
}
