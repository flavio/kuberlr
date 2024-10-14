//go:build linux || darwin
// +build linux darwin

package config

import (
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
)

var configFiles = []string{ //nolint: gochecknoglobals // arrays cannot be go constants
	"/usr/etc/kuberlr.conf",
	"/etc/kuberlr.conf",
	filepath.Join(common.HomeDir(), ".kuberlr", "kuberlr.conf"),
	os.Getenv("KUBERLR_CFG"),
}
