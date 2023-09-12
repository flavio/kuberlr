package common

import "errors"

// NoVersionFoundError error is raised when no kubectl binary
// has yet been downloaded by kuberlr
type NoVersionFoundError struct {
	Err error
}

// Error returns a human description of the error
func (e *NoVersionFoundError) Error() string {
	return "No local kubectl binaries available"
}

// IsNoVersionFound returns true when the given error is of type
// NoVersionFoundError
func IsNoVersionFound(err error) bool {
	var noVersionFoundErr *NoVersionFoundError
	return errors.As(err, &noVersionFoundErr)
}
