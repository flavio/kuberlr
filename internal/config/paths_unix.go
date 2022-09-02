//go:build linux || darwin
// +build linux darwin

package config

import (
	"github.com/flavio/kuberlr/internal/common"
	"path/filepath"
)

var configPaths = []string{
	"/usr/etc/",
	"/etc/",
	filepath.Join(common.HomeDir(), ".kuberlr"),
}
