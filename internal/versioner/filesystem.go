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

// KubectlLocalNamingScheme holds the scheme used to name the kubectl binaries
// downloaded by kuberlr
const KubectlLocalNamingScheme = "kubectl-%d.%d.%d"

// BuildKubectNameFromVersion returns how kuberlr will name the kubectl binary
// with the specified version
func BuildKubectNameFromVersion(v semver.Version) string {
	return fmt.Sprintf(KubectlLocalNamingScheme, v.Major, v.Minor, v.Patch)
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
			KubectlLocalNamingScheme,
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
	semver.Sort(versions)

	if versions.Len() == 0 {
		return versions, &NoVersionFoundError{}
	}
	return versions, nil
}

func (h *localCacheHandler) FindCompatibleKubectlAlreadyDownloaded(requestedVersion semver.Version) (semver.Version, error) {
	versions, err := h.LocalKubectlVersions()
	if err != nil {
		return semver.Version{}, err
	}

	if versions.Len() == 0 {
		return semver.Version{}, &NoVersionFoundError{}
	}

	lowerBound := lowerBoundVersion(requestedVersion)
	upperBound := upperBoundVersion(requestedVersion)
	rangeRule := fmt.Sprintf(">=%s <%s", lowerBound.String(), upperBound.String())

	validRange, err := semver.ParseRange(rangeRule)
	if err != nil {
		return semver.Version{}, err
	}

	for i := len(versions) - 1; i >= 0; i-- {
		if validRange(versions[i]) {
			return versions[i], nil
		}
	}

	return semver.Version{}, &NoVersionFoundError{}
}

func lowerBoundVersion(v semver.Version) semver.Version {
	res := v

	res.Patch = 0
	if v.Minor > 0 {
		res.Minor = v.Minor - 1
	}

	return res
}

func upperBoundVersion(v semver.Version) semver.Version {
	return semver.Version{
		Major: v.Major,
		Minor: v.Minor + 2,
		Patch: 0,
	}
}
