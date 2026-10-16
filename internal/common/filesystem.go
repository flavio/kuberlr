package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// ExpandHome replaces a leading "~" in path with the user home directory.
func ExpandHome(path string) string {
	if path == "~" {
		return HomeDir()
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return filepath.Join(HomeDir(), path[2:])
	}

	return path
}

// LocalRootDir returns the default directory under which kuberlr saves
// the kubectl binaries downloaded from kubernetes' upstream mirror.
func LocalRootDir() string {
	return filepath.Join(HomeDir(), ".kuberlr")
}

// LocalDownloadDir returns the path to where kuberlr saves the kubectl
// binaries for the current platform, below root. An empty root means
// LocalRootDir; a leading "~" is expanded to the user home directory.
func LocalDownloadDir(root string) string {
	if root == "" {
		root = LocalRootDir()
	}
	root = ExpandHome(root)
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(root, platform)
}
