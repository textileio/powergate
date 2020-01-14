package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	newWalletCmd.Flags().StringP("type", "t", "bls", "Specifies the wallet type, either bls or secp256k1. Defaults to bls.")

	rootCmd.AddCommand(newWalletCmd)
}

var newWalletCmd = &cobra.Command{
	Use:   "newWallet",
	Short: "Create a new filecoin wallet",
	Long:  `Create a new filecoin wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		typ, err := cmd.Flags().GetString("type")
		checkErr(err)

		address, err = fcClient.Wallet.NewWallet(ctx, typ)
		checkErr(err)

		Success("Wallet address: %v", address)
	},
}
