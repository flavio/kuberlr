package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/finder"
)

// NewRmCmd creates a new `kuberlr rm` cobra command.
func NewRmCmd() *cobra.Command {
	var prune bool
	var all bool
	var dryRun bool

	//nolint: forbidigo // reporting progress on stdout is the point of this command
	cmd := &cobra.Command{
		Use:          "rm [version to remove]",
		Aliases:      []string{"remove", "delete"},
		Short:        "Remove local kubectl binaries",
		SilenceUsage: true,
		Example: `
  Remove one exact release:
  $ kuberlr rm 1.28.3

  Remove every local release of the 1.28 series:
  $ kuberlr rm 1.28

  Remove every release except the newest patch of each local series:
  $ kuberlr rm --prune

  Remove every local kubectl binary:
  $ kuberlr rm --all

  Add --dry-run to see what a command above would remove, without
  removing anything:
  $ kuberlr rm --dry-run --prune`,
		Args: func(_ *cobra.Command, args []string) error {
			if prune || all {
				if len(args) != 0 {
					return errors.New("--prune and --all do not take a version argument")
				}
				return nil
			}
			if len(args) != 1 {
				return errors.New("kuberlr rm needs one version, or one of --prune, --all; run 'kuberlr rm --help' for examples")
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			var arg string
			if len(args) == 1 {
				arg = args[0]
			}

			kubectlFinder := finder.NewKubectlFinder("", "")

			removed, err := kubectlFinder.Remove(arg, prune, all, dryRun)
			for _, path := range removed {
				if dryRun {
					fmt.Printf("Would remove %s\n", path)
				} else {
					fmt.Printf("Removed %s\n", path)
				}
			}

			if err == nil && len(removed) == 0 {
				fmt.Println("Nothing to remove.")
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&prune, "prune", false, "remove every release except the newest patch of each local series")
	cmd.Flags().BoolVar(&all, "all", false, "remove every local kubectl binary")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be removed, without removing anything")
	cmd.MarkFlagsMutuallyExclusive("prune", "all")

	return cmd
}
