package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
)

func init() {
	storageJobsCmd.AddCommand(storageJobsCancelCmd)
	storageJobsCmd.AddCommand(storageJobsCancelQueuedCmd)
	storageJobsCmd.AddCommand(storageJobsCancelExecutingCmd)
}

var storageJobsCancelCmd = &cobra.Command{
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

var storageJobsCancelQueuedCmd = &cobra.Command{
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

		js, err := c.PowClient.StorageJobs.Queued(ctx)
		c.CheckErr(err)

		for _, j := range js.StorageJobs {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			_, err := c.PowClient.StorageJobs.Cancel(ctx, j.Id)
			c.CheckErr(err)
		}
	},
}

var storageJobsCancelExecutingCmd = &cobra.Command{
	Use:   "cancel-executing",
	Short: "Cancel all executing jobs",
	Long:  "Cancel all executing jobs",
	Args:  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := c.MustAuthCtx(context.Background())

		js, err := c.PowClient.StorageJobs.Executing(ctx)
		c.CheckErr(err)

		for _, j := range js.StorageJobs {
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			_, err := c.PowClient.StorageJobs.Cancel(ctx, j.Id)
			c.CheckErr(err)
		}
	},
}
