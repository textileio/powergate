package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	balanceCmd.Flags().StringP("address", "a", "", "wallet address to check the balance of")

	walletCmd.AddCommand(balanceCmd)
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Print the balance of the specified wallet",
	Long:  `Print the balance of the specified wallet`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		addr := viper.GetString("address")

		if len(addr) == 0 {
			Fatal(errors.New("balance command needs a wallet address"))
		}

		s := spin.New(fmt.Sprintf("%s Checking balance for %s...", "%s", addr))
		s.Start()
		bal, err := fcClient.Wallet.WalletBalance(ctx, addr)
		s.Stop()
		checkErr(err)

		Success("Balance: %v", bal)
	},
}
