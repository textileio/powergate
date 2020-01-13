package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(balanceCmd)
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Print the balance of the specified wallet",
	Long:  `Print the balance of the specified wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		addr, err := cmd.Flags().GetString("address")
		checkErr(err)

		if addr == "" {
			Fatal(errors.New("balance command needs a wallet address"))
		}

		bal, err := fcClient.Wallet.WalletBalance(ctx, addr)
		checkErr(err)

		Success("Balance: %v", bal)
	},
}
