package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(storeCmd)
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch storage process",
	Long:  `Watch storage process`,
	Run: func(cmd *cobra.Command, args []string) {
		// fcClient.Wallet.WalletBalance()
		// addr, err := cmd.Flags().GetString("address")
		// checkErr(err)
		fmt.Println("watch here")
	},
}
