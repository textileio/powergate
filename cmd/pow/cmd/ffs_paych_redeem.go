package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsPaychCmd.AddCommand(ffsPaychRedeemCmd)
}

var ffsPaychRedeemCmd = &cobra.Command{
	Use:   "redeem [addr]",
	Short: "Redeem a payment channel",
	Long:  `Redeem a payment channel`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		_, err := fcClient.FFS.RedeemPayChannel(mustAuthCtx(ctx), args[0])
		checkErr(err)
	},
}
