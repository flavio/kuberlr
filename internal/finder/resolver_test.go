package finder

import (
	"errors"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	t.Run("full version is used verbatim", func(t *testing.T) {
		t.Parallel()

		downloaderMock := NewMockdownloadHelper(t)
		versioner := Versioner{downloader: downloaderMock}

		res, err := versioner.ResolveVersion("v1.19.1")

		require.NoError(t, err)
		assert.Equal(t, semver.MustParse("1.19.1"), res.Version)
		assert.False(t, res.LookedUp)
		assert.NoError(t, res.LookupErr)
	})

	t.Run("pre-release version is used verbatim", func(t *testing.T) {
		t.Parallel()

		downloaderMock := NewMockdownloadHelper(t)
		versioner := Versioner{downloader: downloaderMock}

		res, err := versioner.ResolveVersion("1.36.0-beta.1")

		require.NoError(t, err)
		assert.Equal(t, semver.MustParse("1.36.0-beta.1"), res.Version)
		assert.False(t, res.LookedUp)
	})

	t.Run("major.minor is resolved to latest patch", func(t *testing.T) {
		t.Parallel()

		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(36)).
			Return(semver.MustParse("1.36.4"), nil)
		versioner := Versioner{downloader: downloaderMock}

		res, err := versioner.ResolveVersion("1.36")

		require.NoError(t, err)
		assert.Equal(t, semver.MustParse("1.36.4"), res.Version)
		assert.Equal(t, semver.MustParse("1.36.0"), res.Requested)
		assert.True(t, res.LookedUp)
		assert.NoError(t, res.LookupErr)
	})

	t.Run("major.minor falls back to patch 0 on resolver error", func(t *testing.T) {
		t.Parallel()

		resolverErr := errors.New("boom")
		downloaderMock := NewMockdownloadHelper(t)
		downloaderMock.EXPECT().UpstreamStableVersionForMinor(uint64(1), uint64(99)).
			Return(semver.Version{}, resolverErr)
		versioner := Versioner{downloader: downloaderMock}

		res, err := versioner.ResolveVersion("1.99")

		require.NoError(t, err)
		assert.Equal(t, semver.MustParse("1.99.0"), res.Version)
		assert.True(t, res.LookedUp)
		require.Error(t, res.LookupErr)
	})

	t.Run("invalid input is rejected", func(t *testing.T) {
		t.Parallel()

		downloaderMock := NewMockdownloadHelper(t)
		versioner := Versioner{downloader: downloaderMock}

		_, err := versioner.ResolveVersion("not-a-version")

		require.Error(t, err)
	})
}
