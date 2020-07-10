package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	walletCmd.AddCommand(sendCmd)
}

var sendCmd = &cobra.Command{
	Use:   "send [from address] [to address] [amount]",
	Short: "Send Fil from one address to another",
	Long:  `Send Fil from one address to another`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		from := args[0]
		to := args[1]
		amt, err := strconv.ParseInt(args[2], 10, 64)
		checkErr(err)

		s := spin.New(fmt.Sprintf("%s Sending Fil...", "%s"))
		s.Start()
		err = fcClient.Wallet.SendFil(ctx, from, to, amt)
		s.Stop()
		checkErr(err)

		Success("Done!")
	},
}
