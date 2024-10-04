//go:build linux || darwin
// +build linux darwin

package config

import (
	"github.com/flavio/kuberlr/internal/common"
	"os"
	"path/filepath"
)

var configPaths = []string{ //nolint: gochecknoglobals // arrays cannot be go constants
	"/usr/etc/kuberlr.conf",
	"/etc/kuberlr.conf",
	filepath.Join(common.HomeDir(), ".kuberlr", "kuberlr.conf"),
	os.Getenv("KUBERLR_CFG"),
}
