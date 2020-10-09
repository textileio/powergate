package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsPaychCmd.AddCommand(ffsPaychRedeemCmd)
}

var ffsPaychRedeemCmd = &cobra.Command{
	Use:   "redeem [from] [to] [amount]",
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

		s := spin.New("%s Redeeming payment channel...")
		s.Start()
		err := fcClient.FFS.RedeemPayChannel(authCtx(ctx), args[0])
		s.Stop()
		checkErr(err)

		Success("Redeemed payment channel %v", args[0])
	},
}
