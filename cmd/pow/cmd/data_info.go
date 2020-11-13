package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	dataCmd.AddCommand(dataInfoCmd)
}

var dataInfoCmd = &cobra.Command{
	Use:   "info [optional cid1,cid2,...]",
	Short: "Get information about the current storate state of a cid",
	Long:  `Get information about the current storate state of a cid`,
	Args:  cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		res, err := powClient.Data.CidInfo(mustAuthCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
