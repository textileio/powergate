package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Provides commands to interact with cid storage configs",
	Long:  `Provides commands to interact with cid storage configs`,
}
