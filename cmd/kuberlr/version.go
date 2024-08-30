package main

import (
	"fmt"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/flavio/kuberlr/pkg/kuberlr"
)

// NewVersionCmd creates a new `kuberlr version` cobra command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			//nolint: forbidigo // it's fine to print to stdout
			fmt.Printf("%s\n", kuberlr.CurrentVersion().String())
		},
	}
}
