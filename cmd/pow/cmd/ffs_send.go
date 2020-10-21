package cmd

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCmd.AddCommand(ffsSendCmd)
}

var ffsSendCmd = &cobra.Command{
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

		amount, err := strconv.ParseInt(args[2], 10, 64)
		checkErr(err)

		_, err = fcClient.FFS.SendFil(mustAuthCtx(ctx), from, to, amount)
		checkErr(err)
	},
}
