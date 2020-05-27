package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	ffsCmd.AddCommand(ffsAddrsCmd)
}

var ffsAddrsCmd = &cobra.Command{
	Use:   "addrs",
	Short: "Provides commands to manage wallet addresses",
	Long:  `Provides commands to manage wallet addresses`,
}
