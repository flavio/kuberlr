package main

import (
	"fmt"

	"github.com/spf13/cobra"
	
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/flavio/kuberlr/pkg/kuberlr"
)

// NewVersionCmd creates a new `kuberlr version` cobra command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", kuberlr.CurrentVersion().String())
		},
	}
}
