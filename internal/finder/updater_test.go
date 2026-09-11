package finder

import (
	"errors"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestBin(t *testing.T, version string) KubectlBinary {
	t.Helper()
	return KubectlBinary{
		Path:    "/fake/kubectl" + version,
		Version: semver.MustParse(version),
	}
}

func TestPlanUpdates(t *testing.T) {
	t.Parallel()

	t.Run("empty input produces no actions", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(nil, nil)

		downloaderMock := NewMockdownloadHelper(t)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		assert.Empty(t, actions)
	})

	t.Run("listing local binaries fails", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(nil, errors.New("boom"))

		downloaderMock := NewMockdownloadHelper(t)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		_, err := versioner.PlanUpdates()

		require.Error(t, err)
	})

	t.Run("single series behind upstream needs a download", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{newTestBin(t, "1.28.0")}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(28)).
			Return(semver.MustParse("1.28.11"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 1)
		assert.Equal(t, "1.28", actions[0].Series)
		assert.Equal(t, semver.MustParse("1.28.0"), actions[0].Installed)
		assert.Equal(t, semver.MustParse("1.28.11"), actions[0].Latest)
		assert.True(t, actions[0].NeedsDownload())
	})

	t.Run("single series already up to date needs no download", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{newTestBin(t, "1.29.7")}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(29)).
			Return(semver.MustParse("1.29.7"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 1)
		assert.False(t, actions[0].NeedsDownload())
	})

	t.Run("newest installed patch of a series is used", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{
			newTestBin(t, "1.28.0"),
			newTestBin(t, "1.28.3"),
		}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		// only one upstream lookup for the whole 1.28 series
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(28)).
			Return(semver.MustParse("1.28.5"), nil).Once()

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 1)
		assert.Equal(t, semver.MustParse("1.28.3"), actions[0].Installed)
		assert.True(t, actions[0].NeedsDownload())
	})

	t.Run("multiple series are resolved independently", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{
			newTestBin(t, "1.28.0"),
			newTestBin(t, "1.29.7"),
		}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(28)).
			Return(semver.MustParse("1.28.11"), nil)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(29)).
			Return(semver.MustParse("1.29.7"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 2)
		assert.Equal(t, "1.28", actions[0].Series)
		assert.True(t, actions[0].NeedsDownload())
		assert.Equal(t, "1.29", actions[1].Series)
		assert.False(t, actions[1].NeedsDownload())
	})

	t.Run("series are ordered by version, not by string", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{
			newTestBin(t, "1.10.0"),
			newTestBin(t, "1.28.0"),
			newTestBin(t, "1.9.0"),
		}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(9)).
			Return(semver.MustParse("1.9.0"), nil)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(10)).
			Return(semver.MustParse("1.10.0"), nil)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(28)).
			Return(semver.MustParse("1.28.0"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 3)
		assert.Equal(t, []string{"1.9", "1.10", "1.28"}, []string{
			actions[0].Series, actions[1].Series, actions[2].Series,
		})
	})

	t.Run("resolver error on one series doesn't affect the others", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{
			newTestBin(t, "1.19.0"),
			newTestBin(t, "1.29.7"),
		}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(19)).
			Return(semver.Version{}, errors.New("404"))
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(29)).
			Return(semver.MustParse("1.29.7"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 2)

		assert.Equal(t, "1.19", actions[0].Series)
		require.Error(t, actions[0].Err)
		assert.False(t, actions[0].NeedsDownload())

		assert.Equal(t, "1.29", actions[1].Series)
		require.NoError(t, actions[1].Err)
		assert.False(t, actions[1].NeedsDownload())
	})

	t.Run("installed pre-release is superseded by a stable release", func(t *testing.T) {
		t.Parallel()

		finderMock := NewMockiFinder(t)
		finderMock.EXPECT().LocalKubectlBinaries().Return(KubectlBinaries{newTestBin(t, "1.36.0-beta.1")}, nil)

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(36)).
			Return(semver.MustParse("1.36.0"), nil)

		versioner := Versioner{kFinder: finderMock, downloader: downloaderMock}
		actions, err := versioner.PlanUpdates()

		require.NoError(t, err)
		require.Len(t, actions, 1)
		assert.True(t, actions[0].NeedsDownload())
	})
}

func TestDownloadKubectl(t *testing.T) {
	t.Parallel()

	version := semver.MustParse("1.28.5")

	downloaderMock := NewMockdownloadHelper(t)
	downloaderMock.EXPECT().GetKubectlBinary(version, mock.AnythingOfType("string")).RunAndReturn(
		func(_ semver.Version, destination string) error {
			assert.Contains(t, destination, "kubectl1.28.5")
			return nil
		},
	)

	versioner := Versioner{downloader: downloaderMock}
	path, err := versioner.DownloadKubectl(version)

	require.NoError(t, err)
	assert.Contains(t, path, "kubectl1.28.5")
}
