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
	storageInfoCmd.AddCommand(storageInfoListCmd)
}

var storageInfoListCmd = &cobra.Command{
	Use:   "list [optional cid1,cid2,...]",
	Short: "Returns a list of infomration about all stored cids, filtered by cids if provided.",
	Long:  `Returns a list of infomration about all stored cids, filtered by cids if provided.`,
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

		res, err := powClient.StorageInfo.QueryStorageInfo(mustAuthCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
