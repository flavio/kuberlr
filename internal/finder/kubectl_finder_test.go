package finder

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	defer func() {
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
		}
	}()

	localBins := fakeKubectlBinaries(
		td.FakeHome,
		[]string{"1.4.2"},
		&localKubectlNamer{})
	require.NoError(t, createFakeKubectlBinaries(localBins))

	systemBins := fakeKubectlBinaries(
		td.FakeSysBinPath,
		[]string{"2.1.3"},
		&systemKubectlNamer{})
	require.NoError(t, createFakeKubectlBinaries(systemBins))

	//nolint: gocritic // append returns a different array on purpose, we're joining two arrays
	expected := append(systemBins, localBins...)
	actual := td.Finder.AllKubectlBinaries(true)

	assert.Equal(t, expected, actual, "Expected %+v, got %+v instead", expected, actual)

	for i, expectedBin := range expected {
		assert.Condition(t, func() bool {
			return actual[i].Version.Equals(expectedBin.Version)
		}, "Expected %+v, got %+v instead", expectedBin.Version, actual[i].Version)
		assert.Equal(t, expectedBin.Path, actual[i].Path, "Expected %s, got %s instead", expectedBin.Path, actual[i].Path)
	}
}

func TestLocalKubectlVersionsEmptyCache(t *testing.T) {
	td, err := setupFilesystemTest()
	require.NoError(t, err)
	defer func() {
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
		}
	}()

	bins, err := td.Finder.LocalKubectlBinaries()
	require.NoError(t, err)
	assert.Empty(t, bins)
}

func TestLocalKubectlVersionsDownloadDirNotCreated(t *testing.T) {
	td, err := setupFilesystemTest()
	require.NoError(t, err)
	defer func() {
		if err = teardownFilesystemTest(td); err != nil {
			panic(fmt.Sprintf("Error while tearing down test filesystem: %v\n", err))
		}
	}()

	bins, err := td.Finder.LocalKubectlBinaries()
	require.NoError(t, err)
	assert.Empty(t, bins)
}
