package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fpaCmd)
}

var fpaCmd = &cobra.Command{
	Use:   "fpa",
	Short: "Provides commands to manage fpa",
	Long:  `Provides commands to manage fpa`,
}
