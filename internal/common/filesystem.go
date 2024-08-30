package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SystemPath contains the default path to look for kubectl binaries
// installed system-wide.
const SystemPath = "/usr/bin"

// HomeDirEnvKey returns the name of the environment variable
// that holds the name of the user home directory.
func HomeDirEnvKey() string {
	_, found := os.LookupEnv("HOME")
	if found {
		return "HOME"
	}

	return "USERPROFILE" // windows
}

// HomeDir returns current user home directory.
func HomeDir() string {
	return os.Getenv(HomeDirEnvKey())
}

// LocalDownloadDir return the path to where kuberlr saves
// the kubectl binaries downloaded from kubernetes' upstream mirror.
func LocalDownloadDir() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(
		HomeDir(),
		".kuberlr",
		platform,
	)
}
