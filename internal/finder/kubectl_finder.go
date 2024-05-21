package finder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/osexec"

	"github.com/flavio/kuberlr/internal/common"

	"github.com/blang/semver/v4"
)

// KubectlFinder holds data about where to look the kubectl binaries
type KubectlFinder struct {
	LocalBinaryPath string
	SysBinaryPath   string
}

// NewKubectlFinder returns a properly initialized KubectlFinder object
func NewKubectlFinder(local, sys string) *KubectlFinder {
	if local == "" {
		local = common.LocalDownloadDir()
	}
	if sys == "" {
		sys = common.SystemPath
	}

	return &KubectlFinder{
		LocalBinaryPath: local,
		SysBinaryPath:   sys,
	}
}

// SystemKubectlBinaries returns the list of kubectl binaries that are
// available to all the users of the system
func (f *KubectlFinder) SystemKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(f.SysBinaryPath)
}

// LocalKubectlBinaries returns the list of kubectl binaries that are
// available only to the user currently running kuberlr
func (f *KubectlFinder) LocalKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(f.LocalBinaryPath)
}

// AllKubectlBinaries returns all the kubectl binaries available to the
// user running kuberlr
func (f *KubectlFinder) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	var bins KubectlBinaries

	localBin, err := f.LocalKubectlBinaries()
	if err == nil {
		bins = append(bins, localBin...)
	}

	systemBin, err := f.SystemKubectlBinaries()
	if err == nil {
		bins = append(bins, systemBin...)
	}

	SortKubectlByVersion(bins, reverseSort)

	return bins
}

// FindCompatibleKubectl returns a kubectl binary compatible with the
// version given via the `requestedVersion` parameter
func (f *KubectlFinder) FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error) {
	bins := f.AllKubectlBinaries(true)
	if len(bins) == 0 {
		return KubectlBinary{}, &common.NoVersionFoundError{}
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

	return KubectlBinary{}, &common.NoVersionFoundError{}
}

// MostRecentKubectlAvailable returns the most recent version of
// kubectl available on the system. It could be something downloaded
// by kuberlr or something already available on the system
func (f *KubectlFinder) MostRecentKubectlAvailable() (KubectlBinary, error) {
	bins := f.AllKubectlBinaries(true)

	if len(bins) == 0 {
		return KubectlBinary{}, &common.NoVersionFoundError{}
	}

	return bins[0], nil
}

func inferLocalKubectlVersion(filename string) (semver.Version, error) {
	var major, minor, patch uint64
	numScans, err := fmt.Sscanf(
		osexec.TrimExt(filename),
		common.KubectlLocalNamingScheme,
		&major,
		&minor,
		&patch)

	if numScans == 3 && err == nil {
		sv := semver.Version{
			Major: major,
			Minor: minor,
			Patch: patch,
		}
		return sv, nil
	}
	return semver.Version{}, errors.New("not parsable")
}

func inferSystemKubectlVersion(filename string) (semver.Version, error) {
	var major, minor uint64
	numScans, err := fmt.Sscanf(
		osexec.TrimExt(filename),
		common.KubectlSystemNamingScheme,
		&major,
		&minor)

	if numScans == 2 && err == nil {
		sv := semver.Version{
			Major: major,
			Minor: minor,
			Patch: 0,
		}
		return sv, nil
	}
	return semver.Version{}, errors.New("not parsable")
}

func findKubectlBinaries(path string) (KubectlBinaries, error) {
	var binaries KubectlBinaries

	kubectlBins, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return binaries, nil
		}
		return binaries, err
	}

	for _, file := range kubectlBins {
		var version semver.Version
		var internalErr error

		version, internalErr = inferLocalKubectlVersion(file.Name())
		if internalErr != nil {
			version, internalErr = inferSystemKubectlVersion(file.Name())
			if internalErr != nil {
				continue
			}
		}

		bin := KubectlBinary{
			Path:    filepath.Join(path, file.Name()),
			Version: version,
		}
		binaries = append(binaries, bin)
	}

	return binaries, nil
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
	//nolint: mnd
	return semver.Version{
		Major: v.Major,
		Minor: v.Minor + 2,
		Patch: 0,
	}
}
