package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	ffsAddrsCmd.AddCommand(ffsAddrsListCmd)
}

var ffsAddrsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the wallet adresses for the ffs instance",
	Long:  `List the wallet adresses for the ffs instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.Addrs(mustAuthCtx(ctx))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
