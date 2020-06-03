package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	ffsCmd.AddCommand(ffsPaychCmd)
}

var ffsPaychCmd = &cobra.Command{
	Use:   "paych",
	Short: "Provides commands to manage payment channels",
	Long:  `Provides commands to manage payment channels`,
}
