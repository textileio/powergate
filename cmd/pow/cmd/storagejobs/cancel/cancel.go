package cancel

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "cancel [jobid]",
	Short: "Cancel an executing storage job",
	Long:  `Cancel an executing storage job`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := c.PowClient.StorageJobs.Cancel(c.MustAuthCtx(ctx), args[0])
		c.CheckErr(err)
	},
}
