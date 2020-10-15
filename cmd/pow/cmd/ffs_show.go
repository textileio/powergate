package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func init() {
	ffsShowCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsShowCmd)
}

var ffsShowCmd = &cobra.Command{
	Use:   "show [optional cid]",
	Short: "Show pinned cid data",
	Long:  `Show pinned cid data`,
	Args:  cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		var res protoreflect.ProtoMessage
		if len(args) == 1 {
			c, err := cid.Parse(args[0])
			checkErr(err)
			res, err = fcClient.FFS.Show(authCtx(ctx), c)
			checkErr(err)

		} else {
			var err error
			res, err = fcClient.FFS.ShowAll(authCtx(ctx))
			checkErr(err)
		}

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
