// Package logger owns everything kuberlr prints to the terminal on stderr.
//
// The output is meant to look like the one of a regular CLI tool: every line
// starts with the "kuberlr:" prefix, has no timestamps, PIDs or source
// locations, and uses colors when the terminal supports them.
//
// Messages are emitted through the standard log/slog package. The package
// installs a [slog.Handler] that renders records in that CLI-friendly format.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// LevelTrace is the most detailed logging level. It sits below [slog.LevelDebug]
// and is enabled with the highest verbosity setting.
const LevelTrace = slog.Level(-8)

// Verbosity levels accepted by the "Verbosity" setting.
const (
	// VerbosityDefault shows only informational messages, warnings and errors.
	VerbosityDefault = 0
	// VerbosityDebug adds debug messages, like the reason why the remote
	// kubernetes version could not be detected.
	VerbosityDebug = 1
	// VerbosityTrace adds trace messages, the most detailed level.
	VerbosityTrace = 2
	// VerbosityMax is the highest verbosity value that has an effect.
	VerbosityMax = VerbosityTrace
)

// LevelFromVerbosity maps a numeric verbosity (0, 1, 2) to the minimum
// [slog.Level] that is emitted. Values above VerbosityMax behave like VerbosityMax,
// negative values behave like VerbosityDefault.
func LevelFromVerbosity(verbosity int) slog.Level {
	switch {
	case verbosity <= VerbosityDefault:
		return slog.LevelInfo
	case verbosity == VerbosityDebug:
		return slog.LevelDebug
	default:
		return LevelTrace
	}
}

// Logger is the logger used for all the terminal output of kuberlr. It is a
// [slog.Logger] extended with a trace level, a fatal helper and the knowledge
// of whether an animated progress bar can be shown.
type Logger struct {
	*slog.Logger

	showProgress bool
}

func newLogger(handler slog.Handler, showProgress bool) *Logger {
	return &Logger{
		Logger:       slog.New(handler),
		showProgress: showProgress,
	}
}

// Discard returns a logger that drops every record and never shows a
// progress bar. It is handy as a fallback when no logger has been provided.
func Discard() *Logger {
	return newLogger(slog.DiscardHandler, false)
}

// Trace logs the message at LevelTrace.
func (l *Logger) Trace(msg string, args ...any) {
	l.Log(context.Background(), LevelTrace, msg, args...)
}

// Fatalf logs the formatted message at error level and terminates the
// process with exit code 1.
func (l *Logger) Fatalf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// ShowProgress reports whether an animated progress bar can be shown on
// stderr. It is false when quiet mode is enabled or when stderr is not a
// terminal (e.g. when the output is redirected to a file or captured by CI).
func (l *Logger) ShowProgress() bool {
	return l.showProgress
}

// loggerKey is the context key under which the logger is stored.
type loggerKey struct{}

// NewContext returns a copy of ctx carrying the given logger.
func NewContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

// FromContext returns the logger stored in ctx by NewContext. When ctx carries
// no logger, a logger that discards everything is returned.
func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		if log, ok := ctx.Value(loggerKey{}).(*Logger); ok && log != nil {
			return log
		}
	}
	return Discard()
}
