package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
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

		res, err := fcClient.FFS.CreatePayChannel(mustAuthCtx(ctx), from, to, uint64(amt))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
