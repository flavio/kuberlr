//go:build windows
// +build windows

package config

import (
	"github.com/flavio/kuberlr/internal/common"
	"os"
	"path/filepath"
)

var configPaths = []string{
	filepath.Join(os.Getenv("APPDATA"), "kuberlr", "kuberlr.conf"),
	filepath.Join(os.Getenv("PROGRAMDATA"), "kuberlr", "kuberlr.conf"),
	filepath.Join(common.HomeDir(), ".kuberlr", "kuberlr.conf"),
	os.Getenv("KUBERLR_CFG"),
}
