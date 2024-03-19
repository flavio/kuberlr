//go:build linux || darwin
// +build linux darwin

package config

import (
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
)

var configPaths = []string{ //nolint: gochecknoglobals // arrays cannot be go constants
	"/usr/etc/",
	"/etc/",
	ThisExecutableDir(),
	filepath.Join(common.HomeDir(), ".kuberlr"),
}
