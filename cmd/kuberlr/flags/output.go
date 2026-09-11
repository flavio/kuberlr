// Package flags defines the command line flags shared by the kuberlr subcommands.
package flags

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/flavio/kuberlr/internal/logger"
)

// Flag names.
const (
	VerbosityFlag = "verbosity"
	QuietFlag     = "quiet"
	ColorFlag     = "color"
)

// Output holds the values of the flags that control the terminal output.
type Output struct {
	Verbosity int
	Quiet     bool
	Color     string
}

// Register adds the output flags to the given flag set.
func (o *Output) Register(fs *pflag.FlagSet) {
	fs.IntVarP(&o.Verbosity, VerbosityFlag, "v", logger.VerbosityDefault,
		fmt.Sprintf("verbosity level [%d-%d]: %d shows informational messages, %d adds debug messages, %d adds trace messages",
			logger.VerbosityDefault, logger.VerbosityMax, logger.VerbosityDefault, logger.VerbosityDebug, logger.VerbosityTrace))
	fs.BoolVarP(&o.Quiet, QuietFlag, "q", false,
		"only show warnings and errors, hide informational messages and the download progress bar")
	fs.StringVar(&o.Color, ColorFlag, string(logger.ColorAuto),
		fmt.Sprintf("when to use colors: %s, %s or %s", logger.ColorAuto, logger.ColorAlways, logger.ColorNever))
}

// Apply overlays the flags that were explicitly set on the command line on top
// of the given options, which usually come from the configuration file and the
// environment. Flags always win over configuration.
func (o *Output) Apply(fs *pflag.FlagSet, opts logger.Options) (logger.Options, error) {
	if fs.Changed(VerbosityFlag) {
		opts.Verbosity = o.Verbosity
	}
	if fs.Changed(QuietFlag) {
		opts.Quiet = o.Quiet
	}
	if fs.Changed(ColorFlag) {
		color, err := logger.ParseColorMode(o.Color)
		if err != nil {
			return opts, err
		}
		opts.Color = color
	}
	return opts, nil
}
