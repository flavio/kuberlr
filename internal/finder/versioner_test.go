package finder

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/flavio/kuberlr/internal/common"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTimeoutError struct {
	Err error
}

func (e *mockTimeoutError) Error() string {
	return "mock for timeout error"
}

func (e *mockTimeoutError) Timeout() bool {
	return true
}

func TestEnsureCompatibleKubectlAvailableDownloadsKubectlBinaryWhenNeeded(t *testing.T) {
	tests := []struct {
		name                     string
		kubectlAvailableVersions []string
		requestedVersion         semver.Version
		expectedToMakeDownloads  bool
		downloadAllowed          bool
		expectsError             bool
		useLatestIfNoCompatible  bool
	}{
		{
			name:                     "requested version can be satisfied by already downloaded kubectl binary",
			kubectlAvailableVersions: []string{"1.29.0", "1.26.0"},
			requestedVersion:         semver.MustParse("1.30.2"),
			expectedToMakeDownloads:  false,
			downloadAllowed:          true,
			useLatestIfNoCompatible:  false,
			expectsError:             false,
		},
		{
			name:                     "no kubectl binary available",
			kubectlAvailableVersions: []string{},
			requestedVersion:         semver.MustParse("1.9.0"),
			expectedToMakeDownloads:  true,
			downloadAllowed:          true,
			useLatestIfNoCompatible:  false,
			expectsError:             false,
		},
		{
			name:                     "no compatible kubectl binary available",
			kubectlAvailableVersions: []string{"1.3.0", "1.2.0", "1.1.0"},
			requestedVersion:         semver.MustParse("2.4.0"),
			expectedToMakeDownloads:  true,
			downloadAllowed:          true,
			useLatestIfNoCompatible:  false,
			expectsError:             false,
		},
		{
			name:                     "no compatible kubectl binary available but downloads are not allowed",
			kubectlAvailableVersions: []string{"1.3.0", "1.2.0", "1.1.0"},
			requestedVersion:         semver.MustParse("2.4.0"),
			expectedToMakeDownloads:  false,
			downloadAllowed:          false,
			useLatestIfNoCompatible:  false,
			expectsError:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestedVersion := tt.requestedVersion

			kubectlBins := KubectlBinaries{}
			for _, version := range tt.kubectlAvailableVersions {
				kubectlBins = append(kubectlBins, KubectlBinary{
					Version: semver.MustParse(version),
					Path:    fmt.Sprintf("path/to/kubectl-%s", version),
				})
			}

			finderMock := NewMockiFinder(t)
			finderMock.EXPECT().AllKubectlBinaries(true).Return(kubectlBins)

			downloaderMock := NewMockdownloadHelper(t)
			if tt.expectedToMakeDownloads && tt.downloadAllowed {
				downloaderMock.EXPECT().GetKubectlBinary(requestedVersion, mock.AnythingOfType("string")).RunAndReturn(
					func(_ semver.Version, destination string) error {
						assert.Contains(t, destination, common.LocalDownloadDir())
						return nil
					},
				)
			}

			versioner := Versioner{
				kFinder:                           finderMock,
				downloader:                        downloaderMock,
				preventRecursiveInvocationEnvName: fmt.Sprintf("KUBERLR_RESOLVING_VERSION_%d", rand.Intn(100)),
			}

			_, err := versioner.EnsureCompatibleKubectlAvailable(requestedVersion, tt.downloadAllowed, tt.useLatestIfNoCompatible)
			if tt.expectsError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKubectlVersionToUseTimeoutWhenTalkingWithKubernetesAPIServer(t *testing.T) {
	// a special version used later to indicate that we will not query the latest
	// upstream version
	upstreamVersionDoNotQuery := semver.MustParse("0.0.0")

	tests := []struct {
		name                         string
		kubectlAvailableVersions     []string
		latestUpstreamKubectlVersion semver.Version
		expectedVersion              semver.Version
	}{
		{
			name:                         "using latest kubectl version already downloaded",
			kubectlAvailableVersions:     []string{"1.4.0", "1.2.0"},
			latestUpstreamKubectlVersion: upstreamVersionDoNotQuery,
			expectedVersion:              semver.MustParse("1.4.0"),
		},
		{
			name:                         "use latest upstream version since no kubectl binary has ever been downloaded",
			kubectlAvailableVersions:     []string{},
			latestUpstreamKubectlVersion: semver.MustParse("1.4.0"),
			expectedVersion:              semver.MustParse("1.4.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedVersion := tt.expectedVersion
			kubectlBins := KubectlBinaries{}
			for _, version := range tt.kubectlAvailableVersions {
				kubectlBins = append(kubectlBins, KubectlBinary{
					Version: semver.MustParse(version),
					Path:    fmt.Sprintf("path/to/kubectl-%s", version),
				})
			}

			finderMock := NewMockiFinder(t)
			finderMock.EXPECT().AllKubectlBinaries(true).Return(kubectlBins)

			downloaderMock := NewMockdownloadHelper(t)
			if !tt.latestUpstreamKubectlVersion.EQ(upstreamVersionDoNotQuery) {
				downloaderMock.EXPECT().UpstreamStableVersion().Return(tt.latestUpstreamKubectlVersion, nil)
			}

			expectedTimeout := int64(1)
			apiMock := NewMockkubeAPIHelper(t)
			apiMock.EXPECT().Version(expectedTimeout).Return(semver.Version{}, &mockTimeoutError{})

			versioner := Versioner{
				kFinder:                           finderMock,
				apiServer:                         apiMock,
				downloader:                        downloaderMock,
				preventRecursiveInvocationEnvName: fmt.Sprintf("KUBERLR_RESOLVING_VERSION_%d", rand.Intn(100)),
			}

			actual, err := versioner.KubectlVersionToUse(expectedTimeout)
			require.NoError(t, err)
			assert.Equal(t, expectedVersion, actual, "got %s instead of %s", actual, expectedVersion)
		})
	}
}

func TestKubectlVersionToUseSetsInfiniteRecursionPrevention(t *testing.T) {
	// a special version used later to indicate that we will not query the API
	// server version
	kubeAPIServerVersionDoNotQuery := semver.MustParse("0.0.0")

	tests := []struct {
		name                     string
		recursionHappening       bool
		kubectlAvailableVersions []string
		kubeAPIServerVersion     semver.Version
		expectedVersion          semver.Version
	}{
		{
			name:                     "kuberlr is being called recursively, use most recent kubectl version downloaded",
			recursionHappening:       true,
			kubectlAvailableVersions: []string{"1.4.0", "1.2.0"},
			kubeAPIServerVersion:     kubeAPIServerVersionDoNotQuery,
			expectedVersion:          semver.MustParse("1.4.0"),
		},
		{
			name:                     "no recursion happening, query kube API server",
			recursionHappening:       false,
			kubectlAvailableVersions: []string{"1.4.0", "1.2.0"},
			kubeAPIServerVersion:     semver.MustParse("1.20.0"),
			expectedVersion:          semver.MustParse("1.20.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preventRecursiveInvocationEnvName := fmt.Sprintf("KUBERLR_RESOLVING_VERSION_%d", rand.Intn(100))
			if tt.recursionHappening {
				t.Setenv(preventRecursiveInvocationEnvName, "1")
			}

			expectedVersion := tt.expectedVersion
			kubectlBins := KubectlBinaries{}
			for _, version := range tt.kubectlAvailableVersions {
				kubectlBins = append(kubectlBins, KubectlBinary{
					Version: semver.MustParse(version),
					Path:    fmt.Sprintf("path/to/kubectl-%s", version),
				})
			}

			finderMock := NewMockiFinder(t)
			if tt.recursionHappening {
				finderMock.EXPECT().AllKubectlBinaries(true).Return(kubectlBins)
			}

			// No expectation on downloader, as it should not be called
			downloaderMock := NewMockdownloadHelper(t)

			expectedTimeout := int64(1)
			apiMock := NewMockkubeAPIHelper(t)
			if !tt.recursionHappening {
				apiMock.EXPECT().Version(expectedTimeout).Return(tt.kubeAPIServerVersion, nil)
			}

			versioner := Versioner{
				kFinder:                           finderMock,
				apiServer:                         apiMock,
				downloader:                        downloaderMock,
				preventRecursiveInvocationEnvName: preventRecursiveInvocationEnvName,
			}

			actual, err := versioner.KubectlVersionToUse(expectedTimeout)
			require.NoError(t, err)
			assert.Equal(t, expectedVersion, actual, "got %s instead of %s", actual, expectedVersion)
		})
	}
}

func TestFindCompatibleKubectl(t *testing.T) {
	// a special version used later to indicate that no match is expected
	noVersionExpected := semver.MustParse("0.0.0")

	tests := []struct {
		name                     string
		kubectlAvailableVersions []string
		requestedVersion         semver.Version
		expectedVersion          semver.Version
	}{
		{
			name: "lower bound match",
			kubectlAvailableVersions: []string{
				"3.0.0",
				"2.1.3",
				"1.4.2",
				"1.1.3",
			},
			requestedVersion: semver.MustParse("1.5.13"),
			expectedVersion:  semver.MustParse("1.4.2"),
		},
		{
			name: "upper bound match",
			kubectlAvailableVersions: []string{
				"2.1.3",
				"1.4.2",
				"1.1.3",
			},
			requestedVersion: semver.MustParse("2.1.0"),
			expectedVersion:  semver.MustParse("2.1.3"),
		},
		{
			name: "most recent version",
			kubectlAvailableVersions: []string{
				"2.1.3",
				"1.5.3",
				"1.4.2",
				"1.1.3",
			},
			requestedVersion: semver.MustParse("1.4.0"),
			expectedVersion:  semver.MustParse("1.5.3"),
		},
		{
			name:                     "no kubectl binaries available",
			kubectlAvailableVersions: []string{},
			requestedVersion:         semver.MustParse("1.4.0"),
			expectedVersion:          noVersionExpected,
		},
		{
			name:                     "no compatible version available",
			kubectlAvailableVersions: []string{"1.3.0", "1.2.0", "1.1.0"},
			requestedVersion:         semver.MustParse("2.4.0"),
			expectedVersion:          noVersionExpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubectlBins := KubectlBinaries{}
			for _, version := range tt.kubectlAvailableVersions {
				kubectlBins = append(kubectlBins, KubectlBinary{
					Version: semver.MustParse(version),
					Path:    fmt.Sprintf("path/to/kubectl-%s", version),
				})
			}
			expected := KubectlBinary{
				Version: tt.expectedVersion,
				Path:    fmt.Sprintf("path/to/kubectl-%s", tt.expectedVersion),
			}

			actual, err := findCompatibleKubectl(tt.requestedVersion, kubectlBins)
			if tt.expectedVersion.EQ(noVersionExpected) {
				assert.Error(t, err)
				isNoVersionFound := func() bool {
					return common.IsNoVersionFound(err)
				}
				assert.Condition(t, isNoVersionFound)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expected, actual, "got %s instead of %s", actual, expected)
			}
		})
	}
}

// fallbackTestFinder is a minimal stub that:
// - always fails to find a compatible kubectl,
// - returns a predefined list of local kubectl binaries.
type fallbackTestFinder struct {
	all KubectlBinaries // NOTE: named slice type to match iFinder exactly
}

// Adjust signatures if your real finder interface differs.
func (f *fallbackTestFinder) FindCompatibleKubectl(_ semver.Version) (KubectlBinary, error) {
	return KubectlBinary{}, fmt.Errorf("no compatible kubectl")
}

func (f *fallbackTestFinder) AllKubectlBinaries(reverseSort bool) KubectlBinaries {
	out := make(KubectlBinaries, len(f.all))
	copy(out, f.all)

	// When reverseSort is true, we need newest-first order.
	// Do a simple O(n^2) sort by version descending to avoid importing extra packages.
	if reverseSort && len(out) > 1 {
		for i := 0; i < len(out)-1; i++ {
			for j := i + 1; j < len(out); j++ {
				if out[j].Version.GT(out[i].Version) {
					out[i], out[j] = out[j], out[i]
				}
			}
		}
	}
	return out
}

// newFallbackVersioner wires a Versioner with our fallbackTestFinder.
// NOTE: This assumes Versioner has a field named kFinder and the tests are in the same package.
// If your Versioner must be constructed via a constructor, adapt accordingly.
func newFallbackVersioner(ff *fallbackTestFinder) *Versioner {
	v := &Versioner{}
	// If your field name differs, change this assignment.
	v.kFinder = ff
	return v
}

func fallbackMustVer(t *testing.T, s string) semver.Version {
	t.Helper()
	v, err := semver.Parse(s)
	if err != nil {
		t.Fatalf("bad semver %q: %v", s, err)
	}
	return v
}

// Test: no compatible client, downloads disabled, opt-in enabled -> newest local is used.
func TestEnsureCompatibleKubectlAvailable_FallbackToNewest_WhenOptIn(t *testing.T) {
	t.Parallel()

	ff := &fallbackTestFinder{
		all: KubectlBinaries{
			{Path: "/bin/kubectl-1.27.6", Version: fallbackMustVer(t, "1.27.6")},
			{Path: "/bin/kubectl-1.30.1", Version: fallbackMustVer(t, "1.30.1")}, // newest
			{Path: "/bin/kubectl-1.29.3", Version: fallbackMustVer(t, "1.29.3")},
		},
	}
	v := newFallbackVersioner(ff)

	server := fallbackMustVer(t, "1.24.0")

	got, err := v.EnsureCompatibleKubectlAvailable(
		server,
		/*allowDownload=*/ false,
		/*useLatestIfNoCompatible=*/ true,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/bin/kubectl-1.30.1"
	if got != want {
		t.Fatalf("want newest %s, got %s", want, got)
	}
}

// Test: opt-in enabled but there are no local binaries -> still an error.
func TestEnsureCompatibleKubectlAvailable_Fallback_NoLocals(t *testing.T) {
	t.Parallel()

	ff := &fallbackTestFinder{all: nil}
	v := newFallbackVersioner(ff)

	server := fallbackMustVer(t, "1.24.0")

	if _, err := v.EnsureCompatibleKubectlAvailable(server, false, true); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// Download fails → fallback to newest local kubectl when the option is enabled.
func TestEnsureCompatibleKubectlAvailable_DownloadFails_FallbackToNewest_WhenOptIn(t *testing.T) {
	t.Parallel()

	// Pick a server version that guarantees "no compatible" among the local set below.
	requested := semver.MustParse("2.4.0")

	// IMPORTANT: AllKubectlBinaries(true) is expected to return newest-first.
	kubectlBinsNewestFirst := KubectlBinaries{
		{Version: semver.MustParse("1.30.1"), Path: "path/to/kubectl-1.30.1"}, // newest
		{Version: semver.MustParse("1.29.3"), Path: "path/to/kubectl-1.29.3"},
		{Version: semver.MustParse("1.27.6"), Path: "path/to/kubectl-1.27.6"},
	}

	finderMock := NewMockiFinder(t)
	// Versioner may call AllKubectlBinaries(true) before attempting download (to check compatibility)
	// and again when handling the download error (to pick the newest). Allow two calls.
	finderMock.EXPECT().AllKubectlBinaries(true).Return(kubectlBinsNewestFirst)
	finderMock.EXPECT().AllKubectlBinaries(true).Return(kubectlBinsNewestFirst)

	downloaderMock := NewMockdownloadHelper(t)
	downloaderMock.EXPECT().
		GetKubectlBinary(requested, mock.AnythingOfType("string")).
		Return(fmt.Errorf("network unreachable"))

	versioner := Versioner{
		kFinder:                           finderMock,
		downloader:                        downloaderMock,
		preventRecursiveInvocationEnvName: fmt.Sprintf("KUBERLR_RESOLVING_VERSION_%d", rand.Intn(100)),
	}

	got, err := versioner.EnsureCompatibleKubectlAvailable(
		requested,
		/*allowDownload=*/ true,
		/*useLatestIfNoCompatible=*/ true,
	)
	require.NoError(t, err)
	require.Equal(t, "path/to/kubectl-1.30.1", got)
}
