package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/finder"
	"github.com/flavio/kuberlr/internal/logger"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.FromContext(cmd.Context())
			versioner := finder.NewVersioner(finder.NewKubectlFinder("", ""), log)

			res, err := versioner.ResolveVersion(args[0])
			if err != nil {
				return err
			}

			switch {
			case res.LookupErr != nil:
				log.Warn("could not determine the latest patch release, defaulting to .0",
					"series", fmt.Sprintf("%d.%d", res.Requested.Major, res.Requested.Minor),
					"version", res.Version,
					"error", res.LookupErr)
			case res.LookedUp:
				log.Info("resolved version", "requested", args[0], "version", res.Version)
			}

			_, err = versioner.DownloadKubectl(res.Version)
			return err
		},
	}
}
