package summary

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client/admin"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringP("user-id", "u", "", "optional user id filter to apply")
	Cmd.Flags().StringP("cid", "c", "", "optional cid filter to apply")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "summary",
	Short: "Give a summary of storage jobs in all states",
	Long:  `Give a summary of storage jobs in all states`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		var opts []admin.SummaryOption
		if viper.IsSet("user-id") {
			opts = append(opts, admin.WithUserID(viper.GetString("user-id")))
		}
		if viper.IsSet("cid") {
			opts = append(opts, admin.WithCid(viper.GetString("cid")))
		}

		res, err := c.PowClient.Admin.StorageJobs.Summary(c.AdminAuthCtx(ctx), opts...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
