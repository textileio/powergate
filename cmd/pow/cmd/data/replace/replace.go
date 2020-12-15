package replace

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")
}

// Cmd applies a StorageConfig for c2 equal to that of c1, and removes c1.
var Cmd = &cobra.Command{
	Use:   "replace [cid1] [cid2]",
	Short: "Applies a StorageConfig for c2 equal to that of c1, and removes c1",
	Long:  `Applies a StorageConfig for c2 equal to that of c1, and removes c1. This operation is more efficient than manually removing and adding in two separate operations`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		res, err := c.PowClient.Data.ReplaceData(c.MustAuthCtx(ctx), args[0], args[1])
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))

		if viper.GetBool("watch") {
			c.WatchJobIds(res.JobId)
		}
	},
}
