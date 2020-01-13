package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(asksCmd)
}

var asksCmd = &cobra.Command{
	Use:   "asks",
	Short: "Print the available asks",
	Long:  `Print the available asks`,
	Run: func(cmd *cobra.Command, args []string) {
		// fcClient.Wallet.WalletBalance()
		// addr, err := cmd.Flags().GetString("address")
		// checkErr(err)
		fmt.Println("asks here")
	},
}
