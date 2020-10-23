package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	replaceCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	rootCmd.AddCommand(replaceCmd)
}

var replaceCmd = &cobra.Command{
	Use:   "replace [cid1] [cid2]",
	Short: "Pushes a StorageConfig for c2 equal to that of c1, and removes c1",
	Long:  `Pushes a StorageConfig for c2 equal to that of c1, and removes c1. This operation is more efficient than manually removing and adding in two separate operations`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		res, err := powClient.ReplaceData(mustAuthCtx(ctx), args[0], args[1])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))

		if viper.GetBool("watch") {
			watchJobIds(res.JobId)
		}
	},
}
