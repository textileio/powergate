package remove

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "remove [cid]",
	Short: "Removes a Cid from being tracked as an active storage",
	Long:  `Removes a Cid from being tracked as an active storage. The Cid should have both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		_, err := c.PowClient.StorageConfig.Remove(c.MustAuthCtx(ctx), args[0])
		c.CheckErr(err)
	},
}
