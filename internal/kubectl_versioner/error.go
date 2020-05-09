package kubectl_versioner

type noVersionFound interface {
	NoVersionFound() bool
}

type NoVersionFoundError struct {
	Err error
}

func (e *NoVersionFoundError) Error() string {
	return "No local kubectl binaries available"
}

func (e *NoVersionFoundError) NoVersionFound() bool {
	return true
}
