package finder

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "expected %s to still exist", path)
}

func assertFileGone(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "expected %s to be gone", path)
}

// newRemoverFixture creates a temporary local cache and a temporary system
// directory, populates them with fake kubectl binaries, and registers
// cleanup with t. It returns the finder, the local fixtures, and the system
// fixtures, in that order.
func newRemoverFixture(t *testing.T, local, system []string) (KubectlFinder, KubectlBinaries, KubectlBinaries) {
	t.Helper()

	td, err := setupFilesystemTest()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, teardownFilesystemTest(td))
	})

	localBins := fakeKubectlBinaries(td.FakeHome, local, &localKubectlNamer{})
	require.NoError(t, createFakeKubectlBinaries(localBins))

	systemBins := fakeKubectlBinaries(td.FakeSysBinPath, system, &systemKubectlNamer{})
	require.NoError(t, createFakeKubectlBinaries(systemBins))

	return td.Finder, localBins, systemBins
}

// pathsForVersions returns, in order, the path of the binary in bins whose
// version matches each entry in versions.
func pathsForVersions(t *testing.T, bins KubectlBinaries, versions []string) []string {
	t.Helper()

	if len(versions) == 0 {
		return nil
	}

	paths := make([]string, 0, len(versions))
	for _, v := range versions {
		want := semver.MustParse(v)
		found := false
		for _, b := range bins {
			if b.Version.EQ(want) {
				paths = append(paths, b.Path)
				found = true
				break
			}
		}
		require.True(t, found, "no fixture binary has version %s", v)
	}
	return paths
}

func TestSelectForRemoval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		local   []string
		system  []string
		arg     string
		prune   bool
		all     bool
		want    []string // expected selected versions, in order
		wantErr string   // substring of the expected error; empty means no error
	}{
		{
			name:   "exact version selects only that binary",
			local:  []string{"1.28.0", "1.28.3"},
			system: []string{"1.28.0"},
			arg:    "1.28.3",
			want:   []string{"1.28.3"},
		},
		{
			name:  "major.minor selects every local release of that series",
			local: []string{"1.28.0", "1.28.3", "1.29.0"},
			arg:   "1.28",
			want:  []string{"1.28.0", "1.28.3"},
		},
		{
			name:    "no match anywhere is an error",
			local:   []string{"1.28.0"},
			arg:     "1.99",
			wantErr: "no local kubectl binary matches",
		},
		{
			name:    "a version that exists only system-wide is not found locally",
			system:  []string{"1.28.0"},
			arg:     "1.28",
			wantErr: "no local kubectl binary matches",
		},
		{
			name:    "invalid version is rejected",
			arg:     "not-a-version",
			wantErr: "invalid version",
		},
		{
			name:   "prune keeps only the newest patch of each series",
			local:  []string{"1.28.0", "1.28.3", "1.28.5", "1.29.7"},
			system: []string{"1.28.0"},
			prune:  true,
			want:   []string{"1.28.0", "1.28.3"},
		},
		{
			name:  "prune has nothing to do when every series has one patch",
			local: []string{"1.28.5", "1.29.7"},
			prune: true,
		},
		{
			name:   "all selects every local binary and none of the system ones",
			local:  []string{"1.28.0", "1.29.7"},
			system: []string{"1.30.0"},
			all:    true,
			want:   []string{"1.28.0", "1.29.7"},
		},
		{
			name: "all has nothing to do on an empty cache",
			all:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kf, _, systemBins := newRemoverFixture(t, tt.local, tt.system)

			selected, err := kf.selectForRemoval(tt.arg, tt.prune, tt.all)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Len(t, selected, len(tt.want))
			for i, v := range tt.want {
				assert.True(t, selected[i].Version.EQ(semver.MustParse(v)),
					"selected[%d] = %s, want %s", i, selected[i].Version, v)
			}

			// selectForRemoval must never touch system-wide binaries.
			for _, b := range systemBins {
				assertFileExists(t, b.Path)
			}
		})
	}
}

// TestPruneCandidatesPreRelease exercises pruneCandidates directly instead
// of through the filesystem. The local naming scheme
// (kubectl<major>.<minor>.<patch>) has no room for a pre-release tag, so a
// stable release and its pre-release cannot exist as two distinct files on
// disk.
func TestPruneCandidatesPreRelease(t *testing.T) {
	t.Parallel()

	stable := KubectlBinary{Path: "/fake/stable", Version: semver.MustParse("1.36.0")}
	preRelease := KubectlBinary{Path: "/fake/pre-release", Version: semver.MustParse("1.36.0-beta.1")}

	toRemove := pruneCandidates(KubectlBinaries{preRelease, stable})
	require.Len(t, toRemove, 1)
	assert.Equal(t, preRelease.Path, toRemove[0].Path)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		local       []string
		system      []string
		arg         string
		prune       bool
		all         bool
		dryRun      bool
		wantRemoved []string // local versions expected in the (would-)removed list
		wantErr     string   // substring of the expected error; empty means no error
	}{
		{
			name:        "removes an exact version and leaves the rest",
			local:       []string{"1.28.0", "1.28.3"},
			system:      []string{"1.28.0"},
			arg:         "1.28.0",
			wantRemoved: []string{"1.28.0"},
		},
		{
			name:        "dry run removes nothing",
			local:       []string{"1.28.0"},
			arg:         "1.28.0",
			dryRun:      true,
			wantRemoved: []string{"1.28.0"},
		},
		{
			name:        "all never touches system-wide binaries",
			local:       []string{"1.28.0", "1.29.7"},
			system:      []string{"1.30.0"},
			all:         true,
			wantRemoved: []string{"1.28.0", "1.29.7"},
		},
		{
			name: "nothing to remove returns an empty result and no error",
			all:  true,
		},
		{
			name:    "a version that exists only system-wide removes nothing",
			system:  []string{"1.28.0"},
			arg:     "1.28",
			wantErr: "no local kubectl binary matches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kf, localBins, systemBins := newRemoverFixture(t, tt.local, tt.system)

			removed, err := kf.Remove(tt.arg, tt.prune, tt.all, tt.dryRun)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			wantPaths := pathsForVersions(t, localBins, tt.wantRemoved)
			assert.Equal(t, wantPaths, removed)

			for _, b := range localBins {
				removedFromDisk := !tt.dryRun && slices.Contains(wantPaths, b.Path)
				if removedFromDisk {
					assertFileGone(t, b.Path)
				} else {
					assertFileExists(t, b.Path)
				}
			}

			// Remove must never touch system-wide binaries.
			for _, b := range systemBins {
				assertFileExists(t, b.Path)
			}
		})
	}

	t.Run("removing a symlink to a system binary leaves the target in place", func(t *testing.T) {
		t.Parallel()

		kf, _, systemBins := newRemoverFixture(t, nil, []string{"1.28.0"})

		linkPath := filepath.Join(kf.localBinaryPath, "kubectl1.28.0")
		require.NoError(t, os.Symlink(systemBins[0].Path, linkPath))

		removed, err := kf.Remove("1.28.0", false, false, false)
		require.NoError(t, err)
		assert.Equal(t, []string{linkPath}, removed)

		assertFileGone(t, linkPath)
		assertFileExists(t, systemBins[0].Path)
	})
}
