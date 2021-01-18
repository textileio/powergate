package cancelqueued

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "cancel-queued",
	Short: "Cancel all queued jobs",
	Long:  "Cancel all queued jobs",
	Args:  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := c.MustAuthCtx(context.Background())

		js, err := c.PowClient.StorageJobs.List(ctx, client.ListConfig{Select: client.Queued})
		c.CheckErr(err)

		for _, j := range js.StorageJobs {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			_, err := c.PowClient.StorageJobs.Cancel(ctx, j.Id)
			c.CheckErr(err)
		}
	},
}
