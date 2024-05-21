package main

import (
	"fmt"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/spf13/cobra"
)

// NewGetCmd creates a new `kuberlr get` cobra command
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get [version to get]",
		Short:        "Download the kubectl version specified",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Example: `
  Download version 1.20.0. Note well: the patch version is automatically inferred:
  $ kuberlr get 1.20
  
  Versions can be specified with, or without the 'v' prefix:
  $ kuberlr get v1.19.1`,
		RunE: func(_ *cobra.Command, args []string) error {
			version, err := semver.ParseTolerant(args[0])
			if err != nil {
				return fmt.Errorf("invalid version: %w", err)
			}

			destination := filepath.Join(
				common.LocalDownloadDir(),
				common.BuildKubectlNameForLocalBin(version))

			d := downloader.Downloder{}
			return d.GetKubectlBinary(version, destination)
		},
	}
}
