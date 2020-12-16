package executing

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/cmd/pow/cmd/admin/jobs/util"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	Cmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "executing",
	Short: "List executing storage jobs",
	Long:  `List executing storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.Executing(c.AdminAuthCtx(ctx), util.StorageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
