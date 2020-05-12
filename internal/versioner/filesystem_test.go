package versioner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/flavio/kuberlr/internal/common"
)

type localCacheTestData struct {
	HomeEnv  string
	RealHome string
	TempHome string
}

func setupFilesystemTest() (localCacheTestData, error) {
	homeEnv := common.HomeDirEnvKey()
	realHome := common.HomeDir()

	tempHome, err := ioutil.TempDir("", "kuberlr")
	if err != nil {
		return localCacheTestData{}, err
	}

	os.Setenv(homeEnv, tempHome)

	return localCacheTestData{
		HomeEnv:  homeEnv,
		RealHome: realHome,
		TempHome: tempHome,
	}, nil
}

func teardownFilesystemTest(td localCacheTestData) error {
	os.Setenv(td.HomeEnv, td.RealHome)
	return os.RemoveAll(td.TempHome)
}

func createFakeKubectBin(downloadDir string, version semver.Version) error {
	dest := path.Join(
		downloadDir,
		BuildKubectNameFromVersion(version),
	)

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	return file.Close()
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

	lc := localCacheHandler{}
	actual := lc.LocalDownloadDir()

	if !strings.HasPrefix(actual, td.TempHome) {
		t.Errorf("Expected %s to beging with %s", actual, td.TempHome)
	}
}

func TestLocalKubectlVersions(t *testing.T) {
	td, err := setupFilesystemTest()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}
	defer func() {
		if err := teardownFilesystemTest(td); err != nil {
			fmt.Printf("Error while tearing down test filesystem: %v\n", err)
		}
	}()

	lc := localCacheHandler{}
	err = lc.SetupLocalDirs()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}

	expected := semver.Versions{
		semver.MustParse("1.4.2"),
		semver.MustParse("2.1.3"),
	}

	for _, v := range expected {
		if err := createFakeKubectBin(lc.LocalDownloadDir(), v); err != nil {
			t.Errorf("Unexpected error while creating fake kubectl binary: %+v", err)
		}
	}

	actual, err := lc.LocalKubectlVersions()
	if err != nil {
		t.Errorf("Unexpected error: %+v", err)
	}
	if actual.Len() != expected.Len() {
		t.Errorf("Got %d results instead of %d", actual.Len(), expected.Len())
	}

	semver.Sort(actual)
	semver.Sort(expected)

	for i, expectedV := range expected {
		if !actual[i].Equals(expectedV) {
			t.Errorf("Got %+v instead of %+v", actual[i], expectedV)
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

	lc := localCacheHandler{}
	err = lc.SetupLocalDirs()
	if err != nil {
		t.Errorf("Unexpeted failure: %v", err)
	}

	_, err = lc.LocalKubectlVersions()
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

	lc := localCacheHandler{}

	_, err = lc.LocalKubectlVersions()
	if !isNoVersionFound(err) {
		t.Errorf("Got wrong error type: %T, expected NoVersionFoundError", err)
	}
}
