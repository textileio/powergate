package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dealsCmd)
}

var dealsCmd = &cobra.Command{
	Use:   "deals",
	Short: "Provides commands to view Filecoin deal information",
	Long:  `Provides commands to view Filecoin deal information`,
}
