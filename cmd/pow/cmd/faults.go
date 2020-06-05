package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(faultsCmd)
}

var faultsCmd = &cobra.Command{
	Use:   "faults",
	Short: "Provides commands to view faults data",
	Long:  `Provides commands to view faults data`,
}
