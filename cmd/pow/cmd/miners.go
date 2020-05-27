package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(minersCmd)
}

var minersCmd = &cobra.Command{
	Use:   "miners",
	Short: "Provides commands to view miners data",
	Long:  `Provides commands to view miners data`,
}
