package versioner

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/flavio/kuberlr/internal/common"
)

type localCacheTestData struct {
	HomeEnv        string
	RealHome       string
	TempHome       string
	FakeSysBinPath string
	localCache     localCacheHandler
}

func setupFilesystemTest() (localCacheTestData, error) {
	homeEnv := common.HomeDirEnvKey()
	realHome := common.HomeDir()

	tempHome, err := ioutil.TempDir("", "kuberlr-fake-home")
	if err != nil {
		return localCacheTestData{}, err
	}

	os.Setenv(homeEnv, tempHome)

	fakeSysBin, err := ioutil.TempDir("", "kuberlr-fake-usr-bin")
	if err != nil {
		return localCacheTestData{}, err
	}

	td := localCacheTestData{
		HomeEnv:        homeEnv,
		RealHome:       realHome,
		TempHome:       tempHome,
		FakeSysBinPath: fakeSysBin,
		localCache: localCacheHandler{
			SysBinaryPath: fakeSysBin,
		},
	}
	return td, nil
}

func teardownFilesystemTest(td localCacheTestData) error {
	os.Setenv(td.HomeEnv, td.RealHome)
	err1 := os.RemoveAll(td.TempHome)
	err2 := os.RemoveAll(td.FakeSysBinPath)

	if err1 != nil {
		return err1
	}
	return err2
}

func TestLocalDownloadDir(t *testing.T) {
	td, err := setupFilesystemTest()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	actual := td.localCache.LocalDownloadDir()

	if !strings.HasPrefix(actual, td.TempHome) {
		t.Errorf("Expected %s to beging with %s", actual, td.TempHome)
	}
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
		td.localCache.LocalDownloadDir(),
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
	actual := td.localCache.AllKubectlBinaries(true)

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

	err = td.localCache.SetupLocalDirs()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}

	_, err = td.localCache.LocalKubectlBinaries()
	if !isNoVersionFound(err) {
		t.Errorf("Got wrong error type: %T, expected NoVersionFoundError", err)
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

	_, err = td.localCache.LocalKubectlBinaries()
	if !isNoVersionFound(err) {
		t.Errorf("Got wrong error type: %T, expected NoVersionFoundError", err)
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
		td.localCache.LocalDownloadDir(),
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

	actual, err := td.localCache.FindCompatibleKubectl(semver.MustParse(version))
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
	if !isNoVersionFound(err) {
		t.Errorf("Expected error not found")
	}
}
