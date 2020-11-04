package cmd

import (
	"context"
	"encoding/hex"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	walletCmd.AddCommand(walletSignCmd)
}

var walletSignCmd = &cobra.Command{
	Use:   "sign [hex-encoded-message]",
	Short: "Signs a message with storage profile wallet addresses.",
	Long:  "Signs a message using all wallet addresses associated with the storage profile",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		b, err := hex.DecodeString(args[0])
		checkErr(err)

		res, err := powClient.Wallet.Addresses(mustAuthCtx(ctx))
		checkErr(err)

		data := make([][]string, len(res.Addresses))
		for i, a := range res.Addresses {
			signRes, err := powClient.Wallet.SignMessage(mustAuthCtx(ctx), a.Address, b)
			checkErr(err)
			data[i] = []string{a.Address, hex.EncodeToString(signRes.Signature)}
		}

		RenderTable(os.Stdout, []string{"address", "signature"}, data)
	},
}
