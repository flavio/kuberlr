package finder

import (
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
)

// UpdateAction describes the outcome of checking a single major.minor
// release series against the upstream mirror.
type UpdateAction struct {
	Series    string
	Installed semver.Version
	Latest    semver.Version
	Err       error
}

// NeedsDownload reports whether the release series has a newer patch release
// available upstream than what's currently installed locally.
func (a UpdateAction) NeedsDownload() bool {
	return a.Err == nil && a.Latest.GT(a.Installed)
}

// releaseSeries identifies a kubectl major.minor release series.
type releaseSeries struct {
	Major uint64
	Minor uint64
}

// PlanUpdates lists the local kubectl binaries, groups them by major.minor
// release series and, for each series, asks the upstream mirror what the
// latest known patch release is. Only the newest installed patch of each
// series is considered. The result is sorted by version, ascending.
func (v *Versioner) PlanUpdates() ([]UpdateAction, error) {
	bins, err := v.kFinder.LocalKubectlBinaries()
	if err != nil {
		return nil, fmt.Errorf("failed to list local kubectl binaries: %w", err)
	}

	newest := make(map[releaseSeries]semver.Version)
	for _, b := range bins {
		series := releaseSeries{Major: b.Version.Major, Minor: b.Version.Minor}
		if current, ok := newest[series]; !ok || b.Version.GT(current) {
			newest[series] = b.Version
		}
	}

	installedVersions := make([]semver.Version, 0, len(newest))
	for _, ver := range newest {
		installedVersions = append(installedVersions, ver)
	}
	sort.Slice(installedVersions, func(i, j int) bool {
		return installedVersions[i].LT(installedVersions[j])
	})

	actions := make([]UpdateAction, 0, len(installedVersions))
	for _, installed := range installedVersions {
		latest, resolveErr := v.downloader.UpstreamStableVersionForMinor(installed.Major, installed.Minor)
		actions = append(actions, UpdateAction{
			Series:    fmt.Sprintf("%d.%d", installed.Major, installed.Minor),
			Installed: installed,
			Latest:    latest,
			Err:       resolveErr,
		})
	}

	return actions, nil
}
