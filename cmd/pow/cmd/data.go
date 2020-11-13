package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dataCmd)
}

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Provides commands to interact with general data APIs",
	Long:  `Provides commands to interact with general data APIs`,
}
