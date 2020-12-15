package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	storageJobsCmd.AddCommand(storageJobsLatestFinalCmd)
}

var storageJobsLatestFinalCmd = &cobra.Command{
	Use:   "latest-final [optional cid1,cid2,...]",
	Short: "List the latest final storage jobs",
	Long:  `List the latest final storage jobs`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.StorageJobs.LatestFinal(c.MustAuthCtx(ctx), cids...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
