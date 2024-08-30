package finder

import (
	"fmt"
	"os"
	"testing"
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
			localBinaryPath: fakeHome,
			sysBinaryPath:   fakeSysBin,
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
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
		}
	}()

	localBins := fakeKubectlBinaries(
		td.FakeHome,
		[]string{"1.4.2"},
		&localKubectlNamer{})
	if err = createFakeKubectlBinaries(localBins); err != nil {
		t.Error(err)
	}

	systemBins := fakeKubectlBinaries(
		td.FakeSysBinPath,
		[]string{"2.1.3"},
		&systemKubectlNamer{})
	if err = createFakeKubectlBinaries(systemBins); err != nil {
		t.Error(err)
	}

	//nolint: gocritic // append returns a different array on purpose, we're joining two arrays
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
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
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
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
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
