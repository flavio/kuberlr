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

// KubectlFinder holds data about where to look the kubectl binaries.
type KubectlFinder struct {
	localBinaryPath string
	sysBinaryPath   string
}

// NewKubectlFinder returns a properly initialized KubectlFinder object.
func NewKubectlFinder(local, sys string) *KubectlFinder {
	if local == "" {
		local = common.LocalDownloadDir()
	}
	if sys == "" {
		sys = common.SystemPath
	}

	return &KubectlFinder{
		localBinaryPath: local,
		sysBinaryPath:   sys,
	}
}

// SystemKubectlBinaries returns the list of kubectl binaries that are
// available to all the users of the system.
func (f *KubectlFinder) SystemKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(f.sysBinaryPath)
}

// LocalKubectlBinaries returns the list of kubectl binaries that are
// available only to the user currently running kuberlr.
func (f *KubectlFinder) LocalKubectlBinaries() (KubectlBinaries, error) {
	return findKubectlBinaries(f.localBinaryPath)
}

// AllKubectlBinaries returns all the kubectl binaries available to the
// user running kuberlr.
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
