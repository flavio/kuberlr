package kubectl_versioner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flavio/kuberlr/internal/common"

	"github.com/blang/semver"
)

const KUBECTL_LOCAL_NAMING_SCHEME = "kubectl-%d.%d.%d"

func BuildKubectNameFromVersion(v semver.Version) string {
	return fmt.Sprintf(KUBECTL_LOCAL_NAMING_SCHEME, v.Major, v.Minor, v.Patch)
}

func LocalDownloadDir() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(
		common.HomeDir(),
		".kuberlr",
		platform,
	)
}

func IsKubectlAvailable(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func SetupLocalDirs() error {
	return os.MkdirAll(LocalDownloadDir(), os.ModePerm)
}

func LocalKubectlVersions() (semver.Versions, error) {
	var versions semver.Versions

	kubectlBins, err := ioutil.ReadDir(LocalDownloadDir())
	if err != nil {
		if os.IsNotExist(err) {
			err = &NoVersionFoundError{}
		}
		return versions, err
	}

	for _, f := range kubectlBins {
		var major, minor, patch uint64
		n, err := fmt.Sscanf(
			f.Name(),
			KUBECTL_LOCAL_NAMING_SCHEME,
			&major,
			&minor,
			&patch)

		if n == 3 && err == nil {
			sv := semver.Version{
				Major: major,
				Minor: minor,
				Patch: patch,
			}
			versions = append(versions, sv)
		}
	}

	return versions, nil
}
