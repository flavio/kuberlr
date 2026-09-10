package finder

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
)

// ResolvedVersion is the result of turning a version given by the user into
// the version to download.
type ResolvedVersion struct {
	// Version is the version to download.
	Version semver.Version
	// Requested is the version as the user typed it, padded to
	// major.minor.patch. It equals Version unless the upstream lookup
	// changed the patch level.
	Requested semver.Version
	// LookedUp is true when kuberlr asked the upstream mirror for the
	// latest patch release of the requested major.minor series.
	LookedUp bool
	// LookupErr is set when the lookup failed. When this happens, Version
	// falls back to Requested (patch 0).
	LookupErr error
}

// ResolveVersion parses the version requested by the user. When the user
// only gives a major.minor pair (for example "1.20" or "v1.20"), kuberlr
// asks the upstream mirror what the latest patch release of that series is.
// When the mirror can't answer (for example the series is unknown, or the
// mirror is unreachable), kuberlr falls back to patch release 0.
func (v *Versioner) ResolveVersion(arg string) (ResolvedVersion, error) {
	requested, err := semver.ParseTolerant(arg)
	if err != nil {
		return ResolvedVersion{}, fmt.Errorf("invalid version: %w", err)
	}

	// ParseTolerant pads short versions (for example "1.36" becomes
	// "1.36.0"). If the padded version appears verbatim in the user's
	// input, the patch level was given explicitly and kuberlr downloads
	// exactly that release.
	if strings.Contains(arg, requested.String()) {
		return ResolvedVersion{Version: requested, Requested: requested}, nil
	}

	latest, lookupErr := v.downloader.UpstreamStableVersionForMinor(requested.Major, requested.Minor)
	if lookupErr != nil {
		//nolint:nilerr // the lookup error is reported via ResolvedVersion.LookupErr, not the return error
		return ResolvedVersion{
			Version:   requested,
			Requested: requested,
			LookedUp:  true,
			LookupErr: lookupErr,
		}, nil
	}

	return ResolvedVersion{
		Version:   latest,
		Requested: requested,
		LookedUp:  true,
	}, nil
}
