package cmd

import (
	"context"
	"errors"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsSendCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsSendCmd)
}

var ffsSendCmd = &cobra.Command{
	Use:   "send [from address] [to address] [amount]",
	Short: "Send fil from one managed address to any other address",
	Long:  `Send fil from one managed address to any other address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 3 {
			Fatal(errors.New("you must provide from and to addresses and an amount to send"))
		}

		from := args[0]
		to := args[1]

		amount, err := strconv.ParseInt(args[2], 10, 64)
		checkErr(err)

		s := spin.New("%s Sending fil...")
		s.Start()

		err = fcClient.Ffs.SendFil(authCtx(ctx), from, to, amount)
		s.Stop()
		checkErr(err)

		Success("Sent %v fil from %v to %v", amount, from, to)
	},
}
