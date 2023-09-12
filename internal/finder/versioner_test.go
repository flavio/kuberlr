package finder

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blang/semver/v4"

	"github.com/flavio/kuberlr/internal/common"
)

type mockFinder struct {
	localKubectlBinaries       func() (KubectlBinaries, error)
	systemKubectlBinaries      func() (KubectlBinaries, error)
	findCompatibleKubectl      func(requestedVersion semver.Version) (KubectlBinary, error)
	mostRecentKubectlAvailable func() (KubectlBinary, error)
}

func (m *mockFinder) LocalKubectlBinaries() (KubectlBinaries, error) {
	return m.localKubectlBinaries()
}

func (m *mockFinder) SystemKubectlBinaries() (KubectlBinaries, error) {
	return m.systemKubectlBinaries()
}

func (m *mockFinder) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	local, _ := m.localKubectlBinaries()
	system, _ := m.systemKubectlBinaries()

	//nolint: gocritic
	all := append(local, system...)
	SortKubectlByVersion(all, reverseSort)
	return all
}

func (m *mockFinder) FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error) {
	return m.findCompatibleKubectl(requestedVersion)
}

func (m *mockFinder) MostRecentKubectlAvailable() (KubectlBinary, error) {
	return m.mostRecentKubectlAvailable()
}

type mockDownloader struct {
	getKubectlBinary      func(semver.Version, string) error
	upstreamStableVersion func() (semver.Version, error)
}

func (m *mockDownloader) GetKubectlBinary(version semver.Version, destination string) error {
	return m.getKubectlBinary(version, destination)
}

func (m *mockDownloader) UpstreamStableVersion() (semver.Version, error) {
	return m.upstreamStableVersion()
}

type mockAPIServer struct {
	version func(timeout int64) (semver.Version, error)
}

func (m *mockAPIServer) Version(timeout int64) (semver.Version, error) {
	return m.version(timeout)
}

type mockTimeoutError struct {
	Err error
}

func (e *mockTimeoutError) Error() string {
	return "mock for timeout error"
}

func (e *mockTimeoutError) Timeout() bool {
	return true
}

// keep
func TestEnsureCompatibleKubectlAvailableLocalBinaryFound(t *testing.T) {
	expectedVersion := semver.MustParse("1.9.0")
	expectedPath := "/tmp/kubectl-1.9.0"

	finderMock := mockFinder{}
	finderMock.findCompatibleKubectl = func(v semver.Version) (KubectlBinary, error) {
		return KubectlBinary{
			Version: expectedVersion,
			Path:    expectedPath,
		}, nil
	}

	versioner := Versioner{
		kFinder: &finderMock,
	}

	actual, err := versioner.EnsureCompatibleKubectlAvailable(expectedVersion, true)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if actual != expectedPath {
		t.Errorf("Got %s instead of %s", actual, expectedPath)
	}
}

// keep
func TestEnsureCompatibleKubectlAvailableLocalBinaryNotFound(t *testing.T) {
	finderMock := mockFinder{}
	finderMock.findCompatibleKubectl = func(v semver.Version) (KubectlBinary, error) {
		return KubectlBinary{}, &common.NoVersionFoundError{}
	}

	downloaderInvoked := false
	downloaderMock := mockDownloader{}
	downloaderMock.getKubectlBinary = func(semver.Version, string) error {
		downloaderInvoked = true
		return nil
	}

	versioner := Versioner{
		kFinder:    &finderMock,
		downloader: &downloaderMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.EnsureCompatibleKubectlAvailable(expected, true)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if !strings.HasSuffix(actual, common.BuildKubectlNameForLocalBin(expected)) {
		t.Errorf("Expected filename to end with %s instead I got %s", common.BuildKubectlNameForLocalBin(expected), actual)
	}

	if !downloaderInvoked {
		t.Error("Downloder not used")
	}
}

// keep
func TestKubectlVersionToUseTimeoutButLocalKubectlAvailable(t *testing.T) {
	localBins := fakeKubectlBinaries(
		"/fake/home",
		[]string{"1.2.0", "1.2.3", "1.9.0"},
		&localKubectlNamer{})
	systemBins := KubectlBinaries{}
	expected := localBins[2]

	err := genericTestKubectlVersionToUseTimeout(
		localBins,
		systemBins,
		expected,
		&mockDownloader{})
	if err != nil {
		t.Error(err)
	}
}

// keep
func TestKubectlVersionToUseTimeoutButSystemKubectlAvailable(t *testing.T) {
	systemBins := fakeKubectlBinaries(
		"/usr/bin",
		[]string{"1.2.0", "1.2.3", "1.9.0"},
		&systemKubectlNamer{})
	localBins := KubectlBinaries{}
	expected := systemBins[2]

	err := genericTestKubectlVersionToUseTimeout(
		localBins,
		systemBins,
		expected,
		&mockDownloader{})
	if err != nil {
		t.Error(err)
	}
}

// keep
func TestKubectlVersionToUseTimeoutAndNoKubectlAvailable(t *testing.T) {
	localBins := KubectlBinaries{}
	systemBins := KubectlBinaries{}
	expected := KubectlBinary{
		Version: semver.MustParse("100.100.100"),
		Path:    "fake",
	}

	downloadMock := mockDownloader{}
	downloadMock.upstreamStableVersion = func() (semver.Version, error) {
		return expected.Version, nil
	}

	err := genericTestKubectlVersionToUseTimeout(
		localBins,
		systemBins,
		expected,
		&downloadMock)
	if err != nil {
		t.Error(err)
	}
}

// keep
func genericTestKubectlVersionToUseTimeout(localBins, systemBins KubectlBinaries, expected KubectlBinary, downloader *mockDownloader) error {
	apiMock := mockAPIServer{}
	apiMock.version = func(timeout int64) (semver.Version, error) {
		return semver.Version{}, &mockTimeoutError{}
	}

	finderMock := mockFinder{}
	finderMock.localKubectlBinaries = func() (KubectlBinaries, error) {
		err := &common.NoVersionFoundError{}
		if len(localBins) > 0 {
			err = nil
		}
		return localBins, err
	}
	finderMock.systemKubectlBinaries = func() (KubectlBinaries, error) {
		err := &common.NoVersionFoundError{}
		if len(systemBins) > 0 {
			err = nil
		}
		return systemBins, err
	}
	finderMock.mostRecentKubectlAvailable = func() (KubectlBinary, error) {
		return expected, nil
	}

	versioner := Versioner{
		kFinder:    &finderMock,
		apiServer:  &apiMock,
		downloader: downloader,
	}

	actual, err := versioner.KubectlVersionToUse(1)
	if err != nil {
		return err
	}

	if !actual.Equals(expected.Version) {
		return fmt.Errorf("Got %s instead of %s", actual, expected)
	}

	return nil
}
