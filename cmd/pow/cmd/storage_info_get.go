package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	storageInfoCmd.AddCommand(storageInfoGetCmd)
}

// Cmd is the command.
var storageInfoGetCmd = &cobra.Command{
	Use:   "get [cid]",
	Short: "Returns the information about a stored cid.",
	Long:  `Returns the information about a stored cid.`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.StorageInfo.StorageInfo(mustAuthCtx(ctx), args[0])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.StorageInfo)
		checkErr(err)

		fmt.Println(string(json))
	},
}
