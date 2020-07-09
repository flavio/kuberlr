// +build windows

package osexec

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)



// Exec executes the program referred to by pathname with the given arguments and environment.
// Windows doesn't support the unix-like `execve`, so we try to emulate it as best we can.
//
// Instead, this function will:
// 1. spawn a new child process
// 2. attach to stdin/out/err
// 3. forward signals
// 4. wait for process to exit, and call os.Exit() with the exit code of the child process.
//
// If successful, this function never returns, because the current program will terminate.
func Exec(pathname string, argv []string, env []string) error {
	// Windows doesn't support the unix-like `execve`
	// even the unix-compat layer basically devolves to CreateProcess + ExitProcess

	args := argv
	if len(args) > 0 {
		args = args[1:] // strip off the command name from the argv
	}

	cmd := exec.Command(pathname, args...)
	cmd.Env = env

	// attach stdin/err/out
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// forward signals to child
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()
	go func() {
		for sig := range sigCh {
			if proc := cmd.Process; proc != nil {
				proc.Signal(sig)
			}
		}
	}()

	// Run the child process
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// child process exited with a failure, and we can hopefully grab that exit code
			if exitErr.ProcessState != nil {
				os.Exit(exitErr.ProcessState.ExitCode())
			}
			// assume exit code 1, conventional for errors
			os.Exit(1)
		}
		// probably child process never started, return the error
		return err
	}
	os.Exit(0)
	// never reached
	return nil
}
