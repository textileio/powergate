package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(asksCmd)
}

var asksCmd = &cobra.Command{
	Use:   "asks",
	Short: "Provides commands to view asks data",
	Long:  `Provides commands to view asks data`,
}
