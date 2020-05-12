package common

import (
	"os"
)

func HomeDirEnvKey() string {
	_, found := os.LookupEnv("HOME")
	if found {
		return "HOME"
	}

	return "USERPROFILE" // windows
}

func HomeDir() string {
	return os.Getenv(HomeDirEnvKey())
}
