//go:build linux || darwin
// +build linux darwin

package osexec

import (
	"syscall"
)

// Exec executes the program referred to by pathname with the given arguments and environment.
// On darwin and linux, this uses the 'execve' syscall.  (See: http://man7.org/linux/man-pages/man2/execve.2.html)
//
// In summary, this causes the program that is currently being run by the calling
// process to be replaced with a new program, with newly initialized stack, heap, and
// (initialized and uninitialized) data segments.
//
// If successful, this function never returns because the current program will "terminate" immediately.
func Exec(pathname string, argv []string, env []string) error {
	err := syscall.Exec(pathname, argv, env)
	if err != nil {
		return err
	}
	panic("execve: unexpected return")
}
