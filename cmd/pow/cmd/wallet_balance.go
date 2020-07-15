package cmd

import (
	"context"
	"fmt"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	walletCmd.AddCommand(balanceCmd)
}

var balanceCmd = &cobra.Command{
	Use:   "balance [address]",
	Short: "Print the balance of the specified wallet address",
	Long:  `Print the balance of the specified wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		addr := args[0]

		s := spin.New(fmt.Sprintf("%s Checking balance for %s...", "%s", addr))
		s.Start()
		bal, err := fcClient.Wallet.Balance(ctx, addr)
		s.Stop()
		checkErr(err)

		Success("Balance: %v", bal)
	},
}
