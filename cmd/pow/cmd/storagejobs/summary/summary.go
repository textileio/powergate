package summary

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "summary [optional cid]",
	Short: "Give a summary of storage jobs in all states",
	Long:  `Give a summary of storage jobs in all states`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		var opts []client.SummaryOption
		if len(args) > 0 {
			opts = append(opts, client.WithCid(args[0]))
		}

		res, err := c.PowClient.StorageJobs.Summary(c.MustAuthCtx(ctx), opts...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
