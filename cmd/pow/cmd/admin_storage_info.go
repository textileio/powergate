package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	adminStorageInfoListCmd.Flags().StringSlice("user-ids", nil, "filter results by provided user ids.")
	adminStorageInfoListCmd.Flags().StringSlice("cids", nil, "filter results by provided cids.")

	adminStorageInfoCmd.AddCommand(adminStorageInfoGetCmd)
	adminStorageInfoCmd.AddCommand(adminStorageInfoListCmd)
}

var adminStorageInfoGetCmd = &cobra.Command{
	Use:   "get [user-id] [cid]",
	Short: "Returns the information about a stored cid.",
	Long:  `Returns the information about a stored cid.`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.Admin.StorageInfo.StorageInfo(mustAuthCtx(ctx), args[0], args[1])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.StorageInfo)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminStorageInfoListCmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of information about all stored cids, filtered by user ids and cids if provided.",
	Long:  `Returns a list of information about all stored cids, filtered by user ids and cids if provided.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		userIDs := viper.GetStringSlice("user-ids")
		cids := viper.GetStringSlice("cids")

		res, err := powClient.Admin.StorageInfo.ListStorageInfo(mustAuthCtx(ctx), userIDs, cids)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
