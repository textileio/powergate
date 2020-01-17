package cmd

import (
	"context"
	"errors"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
)

func init() {
	walletCmd.AddCommand(balanceCmd)
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

		s := spin.New("%s Checking wallet balance...")
		s.Start()
		bal, err := fcClient.Wallet.WalletBalance(ctx, addr)
		s.Stop()
		checkErr(err)

		Success("Balance: %v", bal)
	},
}
