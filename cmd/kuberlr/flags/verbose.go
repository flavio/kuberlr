package flags

import (
	"flag"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

const (
	verboseLevelFlagShort = "v"
	verboseLevelFlagLong  = "verbosity"
	verboseLevelFlagUsage = "log level [0-5]. 0 (Only Error and Warning) to 5 (Maximum detail)."
)

// GetVerboseFlagLevel returns verbose flag level.
func GetVerboseFlagLevel() string {
	if f := flag.CommandLine.Lookup(verboseLevelFlagShort); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		return pflagFlag.Value.String()
	}

	return "0"
}

// RegisterVerboseFlag register verbose flag.
func RegisterVerboseFlag(local *pflag.FlagSet) {
	if f := flag.CommandLine.Lookup(verboseLevelFlagShort); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		pflagFlag.Name = verboseLevelFlagLong
		pflagFlag.Usage = verboseLevelFlagUsage
		local.AddFlag(pflagFlag)
	} else {
		klog.Fatalf("failed to find flag in global flagset (flag): %s", verboseLevelFlagShort)
	}
}
