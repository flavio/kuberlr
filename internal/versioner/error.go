package versioner

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

// NoVersionFoundError returns true if the error is a NoVersionFoundError
// instance
func (e *NoVersionFoundError) NoVersionFound() bool {
	return true
}
