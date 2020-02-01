package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reputationCmd)
}

var reputationCmd = &cobra.Command{
	Use:   "reputation",
	Short: "Provides commands to view miner reputation data",
	Long:  `Provides commands to view miner reputation data`,
}
