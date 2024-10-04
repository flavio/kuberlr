//go:build windows
// +build windows

package config

import (
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
)

var configPaths = []string{
	filepath.Join(os.Getenv("APPDATA"), "kuberlr", "kuberlr.conf"),
	filepath.Join(os.Getenv("PROGRAMDATA"), "kuberlr", "kuberlr.conf"),
	filepath.Join(common.HomeDir(), ".kuberlr", "kuberlr.conf"),
	os.Getenv("KUBERLR_CFG"),
}
