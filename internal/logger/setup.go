package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

// ColorMode controls when colors are used.
type ColorMode string

// Accepted values of the "Color" setting.
const (
	// ColorAuto enables colors only when stderr is a terminal and the NO_COLOR
	// environment variable is not set.
	ColorAuto ColorMode = "auto"
	// ColorAlways enables colors unconditionally.
	ColorAlways ColorMode = "always"
	// ColorNever disables colors unconditionally.
	ColorNever ColorMode = "never"
)

// ParseColorMode converts a user supplied string into a ColorMode. The
// comparison is case-insensitive and ignores surrounding whitespace.
func ParseColorMode(s string) (ColorMode, error) {
	switch mode := ColorMode(strings.ToLower(strings.TrimSpace(s))); mode {
	case ColorAuto, ColorAlways, ColorNever:
		return mode, nil
	case "":
		return ColorAuto, nil
	default:
		return ColorAuto, fmt.Errorf("invalid color mode %q, valid values are: %s, %s, %s",
			s, ColorAuto, ColorAlways, ColorNever)
	}
}

// Options configures the terminal output.
type Options struct {
	// Verbosity is the numeric verbosity level, see LevelFromVerbosity.
	Verbosity int
	// Quiet hides informational messages and the download progress bar; only
	// warnings and errors are shown. It takes precedence over Verbosity.
	Quiet bool
	// Color controls when colors are used.
	Color ColorMode
}

// Setup builds the logger used for all the terminal output of kuberlr. It also
// installs it as the default slog logger and routes the messages emitted by
// client-go through it.
//
// It is safe to call Setup more than once; the last call wins.
func Setup(opts Options) *Logger {
	stderrIsTTY := term.IsTerminal(int(os.Stderr.Fd()))
	colorEnabled := ResolveColor(opts.Color, stderrIsTTY)

	level := LevelFromVerbosity(opts.Verbosity)
	if opts.Quiet {
		level = slog.LevelWarn
	}

	handler := NewHandler(os.Stderr, level, colorEnabled)
	log := newLogger(handler, !opts.Quiet && stderrIsTTY)
	slog.SetDefault(log.Logger)
	routeKlog(handler, level)

	// Keep the colors used by other parts of kuberlr (e.g. the tables printed
	// by `kuberlr bins`) consistent with our own output.
	if colorEnabled {
		text.EnableColors()
	} else {
		text.DisableColors()
	}

	return log
}

// ResolveColor decides whether colors are enabled given the requested mode
// and whether stderr is attached to a terminal.
//
// In ColorAuto mode the NO_COLOR convention (https://no-color.org) is honored:
// when the variable is set to a non-empty value colors are disabled.
func ResolveColor(mode ColorMode, stderrIsTTY bool) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		fallthrough
	default:
		return stderrIsTTY && os.Getenv("NO_COLOR") == ""
	}
}
