package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/finder"
	"github.com/flavio/kuberlr/internal/logger"
)

// NewUpdateCmd creates a new `kuberlr update` cobra command.
func NewUpdateCmd() *cobra.Command {
	//nolint: forbidigo // reporting progress on stdout is the point of this command
	return &cobra.Command{
		Use:   "update",
		Short: "Update local kubectl binaries to the latest patch release",
		Args:  cobra.NoArgs,
		Example: `
  For each major.minor version that you have downloaded, kuberlr asks the
  upstream mirror for the latest patch release. If a newer patch release
  exists, kuberlr downloads it. Old patch releases stay on disk:
  $ kuberlr update`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			log := logger.FromContext(cmd.Context())
			kubectlFinder := finder.NewKubectlFinder("", "")
			versioner := finder.NewVersioner(kubectlFinder, log)

			actions, err := versioner.PlanUpdates()
			if err != nil {
				return err
			}

			if len(actions) == 0 {
				fmt.Println("No local kubectl binaries found.")
				return nil
			}

			var failed bool
			for _, a := range actions {
				if a.Err != nil {
					fmt.Printf("Checking %s: installed %s, could not determine latest patch release (%v) -> skipped\n",
						a.Series, a.Installed, a.Err)
					continue
				}

				if !a.NeedsDownload() {
					fmt.Printf("Checking %s: installed %s, latest %s -> up to date\n",
						a.Series, a.Installed, a.Latest)
					continue
				}

				fmt.Printf("Checking %s: installed %s, latest %s -> downloading\n",
					a.Series, a.Installed, a.Latest)

				if _, dlErr := versioner.DownloadKubectl(a.Latest); dlErr != nil {
					log.Error("failed to download kubectl", "version", a.Latest, "error", dlErr)
					failed = true
				}
			}

			if failed {
				return fmt.Errorf("one or more kubectl binaries could not be updated")
			}

			return nil
		},
	}
}
