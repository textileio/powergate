package cmd

import (
	"context"
	"fmt"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	newCmd.Flags().StringP("type", "t", "bls", "Specifies the wallet type, either bls or secp256k1. Defaults to bls.")
	newCmd.Flags().BoolP("switch", "s", false, "Whether or not to switch to the newly created wallet")

	walletCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new filecoin wallet",
	Long:  `Create a new filecoin wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		typ, err := cmd.Flags().GetString("type")
		checkErr(err)

		s := spin.New("%s Creating new wallet...")
		s.Start()
		address, err := fcClient.Wallet.NewWallet(ctx, typ)
		s.Stop()
		checkErr(err)

		Success("Wallet address: %v", address)

		accounts := viper.GetStringSlice("wallets")
		accounts = append(accounts, address)
		viper.Set("wallets", accounts)

		switchAccount, err := cmd.Flags().GetBool("switch")
		checkErr(err)

		if switchAccount {
			viper.Set("address", address)
		}

		err = viper.WriteConfig()
		checkErr(err)

		message := "Wallet list updated"
		if switchAccount {
			message = fmt.Sprintf("%v and default wallet set to %v", message, address)
		}
		Success(message)
	},
}
