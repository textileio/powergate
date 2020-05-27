package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ffsCmd)
}

var ffsCmd = &cobra.Command{
	Use:   "ffs",
	Short: "Provides commands to manage ffs",
	Long:  `Provides commands to manage ffs`,
}
