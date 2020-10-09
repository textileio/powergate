package cmd

import (
	"context"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
)

func init() {
	ffsPaychCmd.AddCommand(ffsPaychCreateCmd)
}

var ffsPaychCreateCmd = &cobra.Command{
	Use:   "create [from] [to] [amount]",
	Short: "Create a payment channel",
	Long:  `Create a payment channel`,
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
		amt, err := strconv.ParseInt(args[2], 10, 64)
		checkErr(err)

		s := spin.New("%s Creating payment channel...")
		s.Start()
		chInfo, msgCid, err := fcClient.FFS.CreatePayChannel(authCtx(ctx), from, to, uint64(amt))
		s.Stop()
		checkErr(err)

		Success("Created payment channel with address %v and message cid %v", chInfo.Addr, util.CidToString(msgCid))
	},
}
