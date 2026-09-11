package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/finder"
)

// NewGetCmd creates a new `kuberlr get` cobra command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get [version to get]",
		Short:        "Download a kubectl version",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Example: `
  Download the latest patch release of 1.20. kuberlr asks the upstream
  mirror which patch release is the newest:
  $ kuberlr get 1.20

  Download one exact release. Give the full version:
  $ kuberlr get 1.20.3

  You can write the version with or without the 'v' prefix:
  $ kuberlr get v1.19.1`,
		RunE: func(_ *cobra.Command, args []string) error {
			versioner := finder.NewVersioner(finder.NewKubectlFinder("", ""))

			res, err := versioner.ResolveVersion(args[0])
			if err != nil {
				return err
			}

			switch {
			case res.LookupErr != nil:
				fmt.Fprintf(os.Stderr,
					"Warning: could not determine latest patch release of %d.%d (%v), defaulting to %d.%d.0\n",
					res.Requested.Major, res.Requested.Minor, res.LookupErr, res.Requested.Major, res.Requested.Minor)
			case res.LookedUp:
				fmt.Fprintf(os.Stderr, "Resolved %s to v%s\n", args[0], res.Version)
			}

			_, err = versioner.DownloadKubectl(res.Version)
			return err
		},
	}
}
