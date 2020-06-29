package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// HomeDirEnvKey returns the name of the environment variable
// that holds the name of the user home directory
func HomeDirEnvKey() string {
	_, found := os.LookupEnv("HOME")
	if found {
		return "HOME"
	}

	return "USERPROFILE" // windows
}

// HomeDir returns current user home directory
func HomeDir() string {
	return os.Getenv(HomeDirEnvKey())
}

func LocalDownloadDir() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(
		HomeDir(),
		".kuberlr",
		platform,
	)
}
