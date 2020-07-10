package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	newCmd.Flags().StringP("type", "t", "bls", "specifies the wallet type, either bls or secp256k1. Defaults to bls.")

	walletCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new filecoin wallet address",
	Long:  `Create a new filecoin wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("wallets", []string{})
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		typ, err := cmd.Flags().GetString("type")
		checkErr(err)

		s := spin.New("%s Creating new wallet address...")
		s.Start()
		address, err := fcClient.Wallet.NewAddress(ctx, typ)
		s.Stop()
		checkErr(err)

		Success("Wallet address: %v", address)
	},
}
