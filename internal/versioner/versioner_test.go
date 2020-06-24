package versioner

import (
	"fmt"
	"testing"

	"github.com/blang/semver"
)

type mockSystem struct {
	localDownloadDir      func() string
	setupLocalDirs        func() error
	localKubectlBinaries  func() (KubectlBinaries, error)
	systemKubectlBinaries func() (KubectlBinaries, error)
	findCompatibleKubectl func(requestedVersion semver.Version) (KubectlBinary, error)
}

func (m *mockSystem) LocalDownloadDir() string {
	return m.localDownloadDir()
}

func (m *mockSystem) SetupLocalDirs() error {
	return m.setupLocalDirs()
}

func (m *mockSystem) LocalKubectlBinaries() (KubectlBinaries, error) {
	return m.localKubectlBinaries()
}

func (m *mockSystem) SystemKubectlBinaries() (KubectlBinaries, error) {
	return m.systemKubectlBinaries()
}

func (m *mockSystem) FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error) {
	return m.findCompatibleKubectl(requestedVersion)
}

func (m *mockSystem) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	local, _ := m.localKubectlBinaries()
	system, _ := m.systemKubectlBinaries()

	all := append(local, system...)
	SortByVersion(all, reverseSort)
	return all
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
	version func() (semver.Version, error)
}

func (m *mockAPIServer) Version() (semver.Version, error) {
	return m.version()
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

func genericTestMostRecentKubectlAvailable(localBins, systemBins KubectlBinaries, expected KubectlBinary) error {
	systemMock := mockSystem{}
	systemMock.localKubectlBinaries = func() (KubectlBinaries, error) {
		bins := localBins
		return bins, nil
	}
	systemMock.systemKubectlBinaries = func() (KubectlBinaries, error) {
		bins := systemBins
		return bins, nil
	}

	versioner := Versioner{
		cache: &systemMock,
	}

	actual, err := versioner.MostRecentKubectlAvailable()
	if err != nil {
		return err
	}

	if !actual.Version.Equals(expected.Version) {
		return fmt.Errorf("Got %s instead of %s", actual, expected)
	}
	if actual.Path != expected.Path {
		return fmt.Errorf("Got %s instead of %s", actual, expected)
	}

	return nil
}

func TestMostRecentKubectlAvailableLocalCacheMoreFreshThanSystem(t *testing.T) {
	localBins := fakeKubectlBinaries(
		"/fake/home",
		[]string{"1.2.0", "1.2.3", "1.9.0"},
		&localKubectlNamer{})
	systemBins := fakeKubectlBinaries(
		"/usr/bin",
		[]string{"1.4.0"},
		&systemKubectlNamer{})
	expected := localBins[2]

	if err := genericTestMostRecentKubectlAvailable(localBins, systemBins, expected); err != nil {
		t.Error(err)
	}
}

func TestMostRecentKubectlAvailableLocalCacheOlderThanSystem(t *testing.T) {
	localBins := fakeKubectlBinaries(
		"/fake/home",
		[]string{"1.2.0", "1.2.3"},
		&localKubectlNamer{})
	systemBins := fakeKubectlBinaries(
		"/usr/bin",
		[]string{"1.4.0"},
		&systemKubectlNamer{})
	expected := systemBins[0]

	if err := genericTestMostRecentKubectlAvailable(localBins, systemBins, expected); err != nil {
		t.Error(err)
	}
}

func TestMostRecentKubectlDownloadedEmptyCache(t *testing.T) {
	systemMock := mockSystem{}
	systemMock.localKubectlBinaries = func() (KubectlBinaries, error) {
		return KubectlBinaries{}, &NoVersionFoundError{}
	}
	systemMock.systemKubectlBinaries = func() (KubectlBinaries, error) {
		return KubectlBinaries{}, &NoVersionFoundError{}
	}

	versioner := Versioner{
		cache: &systemMock,
	}

	_, err := versioner.MostRecentKubectlAvailable()
	if err == nil {
		t.Errorf("Missing error")
	}

	if !isNoVersionFound(err) {
		t.Errorf("Go %T error instead of NoVersionFoundError", err)
	}
}

func TestEnsureCompatibleKubectlAvailableLocalBinaryFound(t *testing.T) {
	expectedVersion := semver.MustParse("1.9.0")
	expectedPath := "/tmp/kubectl-1.9.0"

	systemMock := mockSystem{}
	systemMock.localDownloadDir = func() string { return "/fake" }
	systemMock.findCompatibleKubectl = func(v semver.Version) (KubectlBinary, error) {
		return KubectlBinary{
			Version: expectedVersion,
			Path:    expectedPath,
		}, nil
	}

	versioner := Versioner{
		cache: &systemMock,
	}

	actual, err := versioner.EnsureCompatibleKubectlAvailable(expectedVersion)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if actual != expectedPath {
		t.Errorf("Got %s instead of %s", actual, expectedPath)
	}
}

func TestEnsureCompatibleKubectlAvailableLocalBinaryNotFound(t *testing.T) {
	setupLocalDirsInvoked := false
	systemMock := mockSystem{}
	systemMock.localDownloadDir = func() string { return "/fake" }
	systemMock.findCompatibleKubectl = func(v semver.Version) (KubectlBinary, error) {
		return KubectlBinary{}, &NoVersionFoundError{}
	}

	systemMock.setupLocalDirs = func() error {
		setupLocalDirsInvoked = true
		return nil
	}

	downloaderInvoked := false
	downloaderMock := mockDownloader{}
	downloaderMock.getKubectlBinary = func(semver.Version, string) error {
		downloaderInvoked = true
		return nil
	}

	versioner := Versioner{
		cache:      &systemMock,
		downloader: &downloaderMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.EnsureCompatibleKubectlAvailable(expected)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if actual != versioner.kubectlBinary(expected) {
		t.Errorf("Got %s instead of %s", actual, expected)
	}

	if !downloaderInvoked {
		t.Error("Downloder not used")
	}

	if !setupLocalDirsInvoked {
		t.Error("Local dir setup not done")
	}
}

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

func genericTestKubectlVersionToUseTimeout(localBins, systemBins KubectlBinaries, expected KubectlBinary, downloader *mockDownloader) error {
	apiMock := mockAPIServer{}
	apiMock.version = func() (semver.Version, error) {
		return semver.Version{}, &mockTimeoutError{}
	}

	systemMock := mockSystem{}
	systemMock.localKubectlBinaries = func() (KubectlBinaries, error) {
		err := &NoVersionFoundError{}
		if len(localBins) > 0 {
			err = nil
		}
		return localBins, err
	}
	systemMock.systemKubectlBinaries = func() (KubectlBinaries, error) {
		err := &NoVersionFoundError{}
		if len(systemBins) > 0 {
			err = nil
		}
		return systemBins, err
	}

	versioner := Versioner{
		cache:      &systemMock,
		apiServer:  &apiMock,
		downloader: downloader,
	}

	actual, err := versioner.KubectlVersionToUse()
	if err != nil {
		return err
	}

	if !actual.Equals(expected.Version) {
		return fmt.Errorf("Got %s instead of %s", actual, expected)
	}

	return nil
}
