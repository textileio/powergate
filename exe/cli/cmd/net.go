package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(netCmd)
}

var netCmd = &cobra.Command{
	Use:   "net",
	Short: "Provides commands related to peers and network",
	Long:  `Provides commands related to peers and network`,
}
