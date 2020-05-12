package versioner

import (
	"testing"

	"github.com/blang/semver"
)

type mockLocalCache struct {
	localDownloadDir     func() string
	isKubectlAvailable   func(filename string) bool
	setupLocalDirs       func() error
	localKubectlVersions func() (semver.Versions, error)
}

func (m *mockLocalCache) LocalDownloadDir() string {
	return m.localDownloadDir()
}

func (m *mockLocalCache) IsKubectlAvailable(filename string) bool {
	return m.isKubectlAvailable(filename)
}

func (m *mockLocalCache) SetupLocalDirs() error {
	return m.setupLocalDirs()
}

func (m *mockLocalCache) LocalKubectlVersions() (semver.Versions, error) {
	return m.localKubectlVersions()
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

func TestMostRecentKubectlDownloadedPopulatedCache(t *testing.T) {
	localCacheMock := mockLocalCache{}
	localCacheMock.localKubectlVersions = func() (semver.Versions, error) {
		return semver.Versions{
			semver.MustParse("1.2.0"),
			semver.MustParse("1.9.0"),
			semver.MustParse("1.2.3"),
		}, nil
	}

	versioner := Versioner{
		cache: &localCacheMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.MostRecentKubectlDownloaded()
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if !actual.Equals(expected) {
		t.Errorf("Got %s instead of %s", actual, expected)
	}
}

func TestMostRecentKubectlDownloadedEmptyCache(t *testing.T) {
	localCacheMock := mockLocalCache{}
	localCacheMock.localKubectlVersions = func() (semver.Versions, error) {
		return semver.Versions{}, nil
	}

	versioner := Versioner{
		cache: &localCacheMock,
	}

	_, err := versioner.MostRecentKubectlDownloaded()
	if err == nil {
		t.Errorf("Missing error")
	}

	if !isNoVersionFound(err) {
		t.Errorf("Go %T error instead of NoVersionFoundError", err)
	}
}

func TestEnsureKubectlIsAvailableLocalBinaryFound(t *testing.T) {
	localCacheMock := mockLocalCache{}
	localCacheMock.localDownloadDir = func() string { return "/fake" }
	localCacheMock.isKubectlAvailable = func(f string) bool { return true }

	versioner := Versioner{
		cache: &localCacheMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.EnsureKubectlIsAvailable(expected)
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if actual != versioner.kubectlBinary(expected) {
		t.Errorf("Got %s instead of %s", actual, expected)
	}
}

func TestEnsureKubectlIsAvailableLocalBinaryNotFound(t *testing.T) {
	setupLocalDirsInvoked := false
	localCacheMock := mockLocalCache{}
	localCacheMock.localDownloadDir = func() string { return "/fake" }
	localCacheMock.isKubectlAvailable = func(f string) bool { return false }
	localCacheMock.setupLocalDirs = func() error {
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
		cache:      &localCacheMock,
		downloader: &downloaderMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.EnsureKubectlIsAvailable(expected)
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

func TestKubectlVersionToUseTimeoutButCacheAlreadyPopulated(t *testing.T) {
	apiMock := mockAPIServer{}
	apiMock.version = func() (semver.Version, error) {
		return semver.Version{}, &mockTimeoutError{}
	}

	localCacheMock := mockLocalCache{}
	localCacheMock.localKubectlVersions = func() (semver.Versions, error) {
		return semver.Versions{
			semver.MustParse("1.9.0"),
		}, nil
	}

	versioner := Versioner{
		cache:     &localCacheMock,
		apiServer: &apiMock,
	}

	expected := semver.MustParse("1.9.0")

	actual, err := versioner.KubectlVersionToUse()
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if !actual.Equals(expected) {
		t.Errorf("Got %s instead of %s", actual, expected)
	}
}

func TestKubectlVersionToUseTimeoutAndCacheEmpty(t *testing.T) {
	expected := semver.MustParse("1.9.0")

	apiMock := mockAPIServer{}
	apiMock.version = func() (semver.Version, error) {
		return semver.Version{}, &mockTimeoutError{}
	}

	localCacheMock := mockLocalCache{}
	localCacheMock.localKubectlVersions = func() (semver.Versions, error) {
		return semver.Versions{}, nil
	}

	downloadMock := mockDownloader{}
	downloadMock.upstreamStableVersion = func() (semver.Version, error) {
		return expected, nil
	}

	versioner := Versioner{
		cache:      &localCacheMock,
		apiServer:  &apiMock,
		downloader: &downloadMock,
	}

	actual, err := versioner.KubectlVersionToUse()
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}

	if !actual.Equals(expected) {
		t.Errorf("Got %s instead of %s", actual, expected)
	}
}
