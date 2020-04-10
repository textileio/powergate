package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ffsConfigCmd)
}

var ffsConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Provides commands to manage storage configuration",
	Long:  `Provides commands to manage storage configuration`,
}
