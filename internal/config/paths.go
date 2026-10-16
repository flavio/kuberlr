package config

import (
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
)

// xdgConfigFile returns the path of the configuration file inside the
// XDG base config directory: $XDG_CONFIG_HOME, or .config inside the user
// home directory when that is unset or not absolute, as required by the XDG
// Base Directory Specification. It returns an empty string when neither can
// be determined.
func xdgConfigFile() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(dir) {
		home := common.HomeDir()
		if home == "" {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}

	return filepath.Join(dir, "kuberlr", "kuberlr.conf")
}
