package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/internal/finder"
)

func printBinTable(bins finder.KubectlBinaries) {
	tableWriter := table.NewWriter()
	tableWriter.SetOutputMirror(os.Stdout)
	tableWriter.AppendHeader(table.Row{"#", "Version", "Binary"})
	for i, b := range bins {
		tableWriter.AppendRow([]interface{}{i + 1, b.Version, b.Path})
	}
	tableWriter.Render()
}

// NewBinsCmd creates a new `kuberlr bins` cobra command
func NewBinsCmd() *cobra.Command {
	//nolint: forbidigo
	return &cobra.Command{
		Use:   "bins",
		Short: "Print information about the kubectl binaries found",
		Run: func(_ *cobra.Command, _ []string) {
			kubectlFinder := finder.NewKubectlFinder("", "")
			systemBins, err := kubectlFinder.SystemKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("system-wide kubectl binaries"))
			//nolint: gocritic
			if err != nil {
				fmt.Printf("Error retrieving binaries: %v\n", err)
			} else if len(systemBins) == 0 {
				fmt.Println("No binaries found.")
			} else {
				printBinTable(systemBins)
			}

			fmt.Printf("\n\n")
			localBins, err := kubectlFinder.LocalKubectlBinaries()

			fmt.Printf("%s\n", text.FgGreen.Sprint("local kubectl binaries"))
			//nolint: gocritic
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
