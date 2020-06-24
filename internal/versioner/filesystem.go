package versioner

import (
	"errors"
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

// KubectlSystemNamingScheme holds the scheme used to name the kubectl binaries
// installed system-wide
const KubectlSystemNamingScheme = "kubectl-%d.%d"

// buildKubectlNameForLocalBin returns how kuberlr will name the kubectl binary
// with the specified version when downloading that to the user home
func buildKubectlNameForLocalBin(v semver.Version) string {
	return fmt.Sprintf(KubectlLocalNamingScheme, v.Major, v.Minor, v.Patch)
}

// buildKubectlNameForLocalBin returns how kuberlr expects system-wide
// kubectl binaries to be named
func buildKubectlNameForSystemBin(version semver.Version) string {
	return fmt.Sprintf(KubectlSystemNamingScheme, version.Major, version.Minor)
}

type localCacheHandler struct {
	SysBinaryPath string
}

func NewLocalCacheHandler() *localCacheHandler {
	return &localCacheHandler{
		SysBinaryPath: "/usr/bin",
	}
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

func (h *localCacheHandler) SystemKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(h.SysBinaryPath)
}

func (h *localCacheHandler) LocalKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(h.LocalDownloadDir())
}

func (h *localCacheHandler) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	var bins KubectlBinaries

	localBin, err := h.LocalKubectlBinaries()
	if err == nil {
		bins = append(bins, localBin...)
	}

	systemBin, err := h.SystemKubectlBinaries()
	if err == nil {
		bins = append(bins, systemBin...)
	}

	SortByVersion(bins, reverseSort)

	return bins
}

func inferLocalKubectlVersion(filename string) (semver.Version, error) {
	var major, minor, patch uint64
	n, err := fmt.Sscanf(
		filename,
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
		return sv, nil
	}
	return semver.Version{}, errors.New("Not parsable")
}

func inferSystemKubectlVersion(filename string) (semver.Version, error) {
	var major, minor uint64
	n, err := fmt.Sscanf(
		filename,
		KubectlSystemNamingScheme,
		&major,
		&minor)

	if n == 2 && err == nil {
		sv := semver.Version{
			Major: major,
			Minor: minor,
			Patch: 0,
		}
		return sv, nil
	}
	return semver.Version{}, errors.New("Not parsable")
}

func findKubectlBinaries(path string) (KubectlBinaries, error) {
	var binaries KubectlBinaries

	kubectlBins, err := ioutil.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = &NoVersionFoundError{}
		}
		return binaries, err
	}

	for _, f := range kubectlBins {
		var sv semver.Version
		var err error

		sv, err = inferLocalKubectlVersion(f.Name())
		if err != nil {
			sv, err = inferSystemKubectlVersion(f.Name())
			if err != nil {
				continue
			}
		}

		bin := KubectlBinary{
			Path:    filepath.Join(path, f.Name()),
			Version: sv,
		}
		binaries = append(binaries, bin)
	}

	if len(binaries) == 0 {
		return binaries, &NoVersionFoundError{}
	}
	return binaries, nil
}

func (h *localCacheHandler) FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error) {
	bins := h.AllKubectlBinaries(true)
	if len(bins) == 0 {
		return KubectlBinary{}, &NoVersionFoundError{}
	}

	lowerBound := lowerBoundVersion(requestedVersion)
	upperBound := upperBoundVersion(requestedVersion)
	rangeRule := fmt.Sprintf(">=%s <%s", lowerBound.String(), upperBound.String())

	validRange, err := semver.ParseRange(rangeRule)
	if err != nil {
		return KubectlBinary{}, err
	}

	for _, b := range bins {
		if validRange(b.Version) {
			return b, nil
		}
	}

	return KubectlBinary{}, &NoVersionFoundError{}
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
