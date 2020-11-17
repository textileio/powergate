package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	adminDataCmd.AddCommand(
		adminDataGCStaged,
		adminDataPinnedCids,
	)
}

var adminDataGCStaged = &cobra.Command{
	Use:   "gcstaged",
	Short: "Unpins unused staged data.",
	Long:  `Unpins staged data not used by queued or executing jobs.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.Admin.Data.GCStaged(ctx)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminDataPinnedCids = &cobra.Command{
	Use:   "pinnedCids",
	Short: "List pinned cids information in hot-storage.",
	Long:  "List pinned cids information in hot-storage.",
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.Admin.Data.PinnedCids(ctx)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
