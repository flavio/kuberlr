package versioner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flavio/kuberlr/internal/common"

	"github.com/blang/semver"
)

// KUBECTL_LOCAL_NAMING_SCHEME holds the scheme used to name the kubectl binaries
// downloaded by kuberlr
const KUBECTL_LOCAL_NAMING_SCHEME = "kubectl-%d.%d.%d"

// BuildKubectNameFromVersion returns how kuberlr will name the kubectl binary
// with the specified version
func BuildKubectNameFromVersion(v semver.Version) string {
	return fmt.Sprintf(KUBECTL_LOCAL_NAMING_SCHEME, v.Major, v.Minor, v.Patch)
}

type localCacheHandler struct {
}

func (*localCacheHandler) LocalDownloadDir() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(
		common.HomeDir(),
		".kuberlr",
		platform,
	)
}

func (*localCacheHandler) IsKubectlAvailable(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func (h *localCacheHandler) SetupLocalDirs() error {
	return os.MkdirAll(h.LocalDownloadDir(), os.ModePerm)
}

func (h *localCacheHandler) LocalKubectlVersions() (semver.Versions, error) {
	var versions semver.Versions

	kubectlBins, err := ioutil.ReadDir(h.LocalDownloadDir())
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

	if versions.Len() == 0 {
		return versions, &NoVersionFoundError{}
	}

	return versions, nil
}
