package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(slashingCmd)
}

var slashingCmd = &cobra.Command{
	Use:   "slashing",
	Short: "Provides commands to view slashing data",
	Long:  `Provides commands to view slashing data`,
}
