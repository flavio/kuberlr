package common

type noVersionFound interface {
	NoVersionFound() bool
}

// NoVersionFoundError error is raised when no kubectl binary
// has yet been downloaded by kuberlr
type NoVersionFoundError struct {
	Err error
}

// Error returns a human description of the error
func (e *NoVersionFoundError) Error() string {
	return "No local kubectl binaries available"
}

// NoVersionFound returns true if the error is a NoVersionFoundError instance
func (e *NoVersionFoundError) NoVersionFound() bool {
	return true
}

// IsNoVersionFound returns true when the given error is of type
// NoVersionFoundError
func IsNoVersionFound(err error) bool {
	t, ok := err.(noVersionFound)
	return ok && t.NoVersionFound()
}
