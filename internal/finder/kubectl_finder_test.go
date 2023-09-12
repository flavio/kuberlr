package finder

import (
	"fmt"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/flavio/kuberlr/internal/common"
)

type localCacheTestData struct {
	FakeHome       string
	FakeSysBinPath string
	Finder         KubectlFinder
}

func setupFilesystemTest() (localCacheTestData, error) {
	fakeHome, err := os.MkdirTemp("", "kuberlr-fake-home")
	if err != nil {
		return localCacheTestData{}, err
	}

	fakeSysBin, err := os.MkdirTemp("", "kuberlr-fake-usr-bin")
	if err != nil {
		return localCacheTestData{}, err
	}

	td := localCacheTestData{
		FakeHome:       fakeHome,
		FakeSysBinPath: fakeSysBin,
		Finder: KubectlFinder{
			LocalBinaryPath: fakeHome,
			SysBinaryPath:   fakeSysBin,
		},
	}
	return td, nil
}

func teardownFilesystemTest(td localCacheTestData) error {
	err1 := os.RemoveAll(td.FakeHome)
	err2 := os.RemoveAll(td.FakeSysBinPath)

	if err1 != nil {
		return err1
	}
	return err2
}

func TestAllKubectlBinaries(t *testing.T) {
	td, err := setupFilesystemTest()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	localBins := fakeKubectlBinaries(
		td.FakeHome,
		[]string{"1.4.2"},
		&localKubectlNamer{})
	if err := createFakeKubectlBinaries(localBins); err != nil {
		t.Error(err)
	}

	systemBins := fakeKubectlBinaries(
		td.FakeSysBinPath,
		[]string{"2.1.3"},
		&systemKubectlNamer{})
	if err := createFakeKubectlBinaries(systemBins); err != nil {
		t.Error(err)
	}

	expected := append(systemBins, localBins...)
	actual := td.Finder.AllKubectlBinaries(true)

	if len(expected) != len(actual) {
		t.Errorf("Expected %+v, got %+v instead", expected, actual)
	}

	for i, expectedBin := range expected {
		if !actual[i].Version.Equals(expectedBin.Version) {
			t.Errorf("Got %+v instead of %+v", actual[i].Version, expectedBin.Version)
		}
		if actual[i].Path != expectedBin.Path {
			t.Errorf("Got %+v instead of %+v", actual[i].Path, expectedBin.Path)
		}
	}
}

func TestLocalKubectlVersionsEmptyCache(t *testing.T) {
	td, err := setupFilesystemTest()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	bins, err := td.Finder.LocalKubectlBinaries()
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
	}
	if len(bins) != 0 {
		t.Errorf("Expected empty list")
	}
}

func TestLocalKubectlVersionsDownloadDirNotCreated(t *testing.T) {
	td, err := setupFilesystemTest()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	bins, err := td.Finder.LocalKubectlBinaries()
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
	}
	if len(bins) != 0 {
		t.Errorf("Expected empty list")
	}
}

func findCompatibleKubectlTester(version string, localVersions, systemVersions []string, expected string) error {
	td, err := setupFilesystemTest()
	if err != nil {
		return err
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	localBins := fakeKubectlBinaries(
		td.FakeHome,
		localVersions,
		&localKubectlNamer{})
	if err := createFakeKubectlBinaries(localBins); err != nil {
		return err
	}

	systemBins := fakeKubectlBinaries(
		td.FakeSysBinPath,
		systemVersions,
		&systemKubectlNamer{})
	if err := createFakeKubectlBinaries(systemBins); err != nil {
		return err
	}

	actual, err := td.Finder.FindCompatibleKubectl(semver.MustParse(version))
	if err != nil {
		return err
	}
	expectedVersion := semver.MustParse(expected)
	if !actual.Version.Equals(expectedVersion) {
		return fmt.Errorf("Got %v instead of %v", actual.Version, expectedVersion)
	}

	return nil
}

func TestFindCompatibleKubectlLowerBoundMatchInsideLocalCache(t *testing.T) {
	localVersions := []string{"1.4.2", "2.1.3"}
	systemVersions := []string{"1.1.3"}
	expectedVersion := "1.4.2"

	err := findCompatibleKubectlTester("1.5.13", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlLowerBoundMatchInsideSystem(t *testing.T) {
	localVersions := []string{"1.1.3"}
	systemVersions := []string{"1.4.2", "2.1.3"}
	expectedVersion := "1.4.0"

	err := findCompatibleKubectlTester("1.5.13", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlUpperBoundMatchInsideLocalCache(t *testing.T) {
	localVersions := []string{"1.4.2", "2.1.3"}
	systemVersions := []string{"1.1.3"}
	expectedVersion := "2.1.3"

	err := findCompatibleKubectlTester("2.1.0", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlUpperBoundMatchInsideSystem(t *testing.T) {
	localVersions := []string{"1.1.3"}
	systemVersions := []string{"1.4.2", "2.1.3"}
	expectedVersion := "2.1.0"

	err := findCompatibleKubectlTester("2.1.0", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlMostRecentCompatibleVersionInsideLocalCache(t *testing.T) {
	localVersions := []string{"1.5.3"}
	systemVersions := []string{"1.4.2", "2.1.3"}
	expectedVersion := "1.5.3"

	err := findCompatibleKubectlTester("1.4.0", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlMostRecentCompatibleVersionInsideSystem(t *testing.T) {
	localVersions := []string{"1.4.2", "2.1.3"}
	systemVersions := []string{"1.5.3"}
	expectedVersion := "1.5.0"

	err := findCompatibleKubectlTester("1.4.0", localVersions, systemVersions, expectedVersion)
	if err != nil {
		t.Error(err)
	}
}

func TestFindCompatibleKubectlNoMatchFound(t *testing.T) {
	localVersions := []string{}
	systemVersions := []string{}
	expectedVersion := "0.0.0"

	err := findCompatibleKubectlTester("1.4.0", localVersions, systemVersions, expectedVersion)
	if !common.IsNoVersionFound(err) {
		t.Errorf("Expected error not found")
	}
}
