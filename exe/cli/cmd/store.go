package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store data in filecoin",
	Long:  `Store data in filecoin`,
	Run: func(cmd *cobra.Command, args []string) {
		// fcClient.Wallet.WalletBalance()
		// addr, err := cmd.Flags().GetString("address")
		// checkErr(err)
		fmt.Println("store here")
	},
}
