//go:build windows
// +build windows

package config

import (
	"github.com/flavio/kuberlr/internal/common"
	"os"
	"path/filepath"
)

var configPaths = []string{
	filepath.Join(os.Getenv("APPDATA"), "kuberlr"),
	filepath.Join(os.Getenv("PROGRAMDATA"), "kuberlr"),
	ThisExecutableDir(),
	filepath.Join(common.HomeDir(), ".kuberlr"),
}
