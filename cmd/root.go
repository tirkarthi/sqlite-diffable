package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sqlite-diffable",
	Short: "A command line utility to dump sqlite database as JSON.",
	Long:  `A command line utility to dump sqlite database as JSON. It dumps the metadata of the table including name and schema as JSON. The program also dumps the rows of given table as JSON.`,
}

// Execute command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
