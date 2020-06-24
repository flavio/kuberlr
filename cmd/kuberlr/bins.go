package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	versioner "github.com/flavio/kuberlr/internal/versioner"
)

func printBinTable(bins versioner.KubectlBinaries) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Version", "Binary"})
	for i, b := range bins {
		t.AppendRow([]interface{}{i + 1, b.Version, b.Path})
	}
	t.Render()
}

// NewBinsCmd creates a new `kuberlr bins` cobra command
func NewBinsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bins",
		Short: "Print information about the kubectl binaries found",
		Run: func(cmd *cobra.Command, args []string) {
			localCache := versioner.NewLocalCacheHandler()

			systemBins, err := localCache.SystemKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("system-wide kubectl binaries"))
			if err != nil {
				fmt.Printf("Error retrieving binaries: %v\n", err)
			} else if len(systemBins) == 0 {
				fmt.Println("No binaries found.")
			} else {
				printBinTable(systemBins)
			}

			fmt.Printf("\n\n")
			localBins, err := localCache.LocalKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("local kubectl binaries"))
			if err != nil {
				fmt.Printf("Error retrieving binaries: %v\n", err)
			} else if len(localBins) == 0 {
				fmt.Println("No binaries found.")
			} else {
				printBinTable(localBins)
			}
		},
	}
}
