package cmd

import (
	"context"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	walletCmd.AddCommand(walletSendCmd)
}

var walletSendCmd = &cobra.Command{
	Use:   "send [from address] [to address] [amount]",
	Short: "Send fil from one managed address to any other address",
	Long:  `Send fil from one managed address to any other address`,
	Args:  cobra.ExactArgs(3),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		from := args[0]
		to := args[1]

		amount, ok := new(big.Int).SetString(args[2], 10)
		if !ok {
			checkErr(fmt.Errorf("parsing amount %v", args[2]))
		}

		_, err := powClient.Wallet.SendFil(mustAuthCtx(ctx), from, to, amount)
		checkErr(err)
	},
}
