package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Provides commands about filecoin wallets",
	Long:  `Provides commands about filecoin wallets`,
}
