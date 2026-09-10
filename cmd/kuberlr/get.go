package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"
)

// minorVersionResolver resolves the latest stable patch release of a given
// major.minor release line. It is implemented by *downloader.Downloder and
// is only defined here to allow the resolution logic to be unit tested
// without performing network calls.
type minorVersionResolver interface {
	UpstreamStableVersionForMinor(major, minor uint64) (semver.Version, error)
}

// NewGetCmd creates a new `kuberlr get` cobra command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get [version to get]",
		Short:        "Download the kubectl version specified",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Example: `
  Download the latest patch release of the 1.20 release line. Kuberlr asks
  the upstream kubernetes mirror which patch release is the most recent one:
  $ kuberlr get 1.20

  Download a specific patch release by giving the full version number:
  $ kuberlr get 1.20.3

  Versions can be specified with, or without the 'v' prefix:
  $ kuberlr get v1.19.1`,
		RunE: func(_ *cobra.Command, args []string) error {
			d := downloader.Downloder{}

			version, err := resolveRequestedVersion(args[0], &d)
			if err != nil {
				return err
			}

			destination := filepath.Join(
				common.LocalDownloadDir(),
				common.BuildKubectlNameForLocalBin(version))

			return d.GetKubectlBinary(version, destination)
		},
	}
}

// resolveRequestedVersion parses the version requested by the user. When the
// user only provides a major.minor pair (e.g. "1.20" or "v1.20"), the
// resolver is queried to find out what the most recent patch release of that
// release line is. When the resolver can't answer (e.g. the release line is
// unknown to the upstream mirror, or the mirror is unreachable), kuberlr
// falls back to patch release 0, which was the previous behavior.
func resolveRequestedVersion(arg string, resolver minorVersionResolver) (semver.Version, error) {
	version, err := semver.ParseTolerant(arg)
	if err != nil {
		return semver.Version{}, fmt.Errorf("invalid version: %w", err)
	}

	// ParseTolerant pads short versions (e.g. "1.36" becomes "1.36.0"). If
	// the padded version appears verbatim in the user's input, the patch
	// level was given explicitly and we download exactly that release.
	if strings.Contains(arg, version.String()) {
		return version, nil
	}

	latest, err := resolver.UpstreamStableVersionForMinor(version.Major, version.Minor)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Warning: could not determine latest patch release of %d.%d (%v), defaulting to %d.%d.0\n",
			version.Major, version.Minor, err, version.Major, version.Minor)
		return version, nil
	}

	fmt.Fprintf(os.Stderr, "Resolved %s to v%s\n", arg, latest.String())
	return latest, nil
}
