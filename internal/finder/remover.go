package finder

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver/v4"
)

// versionMatch decides whether a kubectl binary matches the version
// argument given by the user to `kuberlr rm`.
type versionMatch struct {
	version semver.Version
	exact   bool
}

// parseVersionMatch parses arg the same way ResolveVersion does: a full
// version (for example "1.28.3") matches one exact release, a major.minor
// pair (for example "1.28") matches every release of that series.
func parseVersionMatch(arg string) (versionMatch, error) {
	version, err := semver.ParseTolerant(arg)
	if err != nil {
		return versionMatch{}, fmt.Errorf("invalid version: %w", err)
	}

	// ParseTolerant pads short versions (for example "1.28" becomes
	// "1.28.0"). If the padded version appears verbatim in the user's
	// input, the patch level was given explicitly.
	exact := strings.Contains(arg, version.String())

	return versionMatch{version: version, exact: exact}, nil
}

func (m versionMatch) matches(v semver.Version) bool {
	if m.exact {
		return v.EQ(m.version)
	}
	return v.Major == m.version.Major && v.Minor == m.version.Minor
}

// filter returns the binaries in bins that match m.
func (m versionMatch) filter(bins KubectlBinaries) KubectlBinaries {
	var selected KubectlBinaries
	for _, b := range bins {
		if m.matches(b.Version) {
			selected = append(selected, b)
		}
	}
	return selected
}

// Remove removes the local kubectl binaries selected by arg, prune, or all,
// and returns the paths it removed. When dryRun is true, it returns the
// paths it would remove and leaves the filesystem alone. Remove only ever
// looks inside the local kubectl cache, so it never removes a system-wide
// binary.
//
// The caller must set exactly one of: arg (a version or major.minor
// series), prune, or all.
//
//   - arg selects one exact release, or every release of a series.
//   - prune selects every release except the newest patch of each local
//     series.
//   - all selects every local release.
func (f *KubectlFinder) Remove(arg string, prune, all, dryRun bool) ([]string, error) {
	bins, err := f.selectForRemoval(arg, prune, all)
	if err != nil {
		return nil, err
	}

	var removed []string
	var errs []error

	for _, b := range bins {
		if !dryRun {
			if rmErr := os.Remove(b.Path); rmErr != nil {
				errs = append(errs, fmt.Errorf("failed to remove %s: %w", b.Path, rmErr))
				continue
			}
		}

		removed = append(removed, b.Path)
	}

	return removed, errors.Join(errs...)
}

// selectForRemoval lists the local kubectl binaries that a `kuberlr rm`
// request selects. It only reads the local kubectl cache, so it never
// returns a system-wide binary.
func (f *KubectlFinder) selectForRemoval(arg string, prune, all bool) (KubectlBinaries, error) {
	local, err := f.LocalKubectlBinaries()
	if err != nil {
		return nil, fmt.Errorf("failed to list local kubectl binaries: %w", err)
	}

	switch {
	case all:
		return local, nil
	case prune:
		return pruneCandidates(local), nil
	default:
		return matchLocalVersion(arg, local)
	}
}

// matchLocalVersion returns the local binaries matching arg.
func matchLocalVersion(arg string, local KubectlBinaries) (KubectlBinaries, error) {
	match, err := parseVersionMatch(arg)
	if err != nil {
		return nil, err
	}

	if selected := match.filter(local); len(selected) > 0 {
		return selected, nil
	}

	return nil, fmt.Errorf(
		"no local kubectl binary matches %s (kuberlr rm only removes binaries downloaded by kuberlr, run 'kuberlr bins' to see them)",
		arg)
}

// pruneCandidates returns, for every major.minor series found in bins,
// every binary except the newest patch.
func pruneCandidates(bins KubectlBinaries) KubectlBinaries {
	groups := make(map[releaseSeries]KubectlBinaries)
	for _, b := range bins {
		series := releaseSeries{Major: b.Version.Major, Minor: b.Version.Minor}
		groups[series] = append(groups[series], b)
	}

	var toRemove KubectlBinaries
	for _, group := range groups {
		if len(group) < 2 { //nolint:mnd // a group with one binary has nothing to prune
			continue
		}

		newest := group[0]
		for _, b := range group[1:] {
			if b.Version.GT(newest.Version) {
				newest = b
			}
		}

		for _, b := range group {
			if b.Path != newest.Path {
				toRemove = append(toRemove, b)
			}
		}
	}

	SortKubectlByVersion(toRemove, false)

	return toRemove
}
