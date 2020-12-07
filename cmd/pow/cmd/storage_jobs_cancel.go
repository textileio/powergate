package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := powClient.StorageJobs.Cancel(mustAuthCtx(ctx), args[0])
		checkErr(err)
	},
}

var storageJobsCancelQueuedCmd = &cobra.Command{
	Use:   "cancel-queued",
	Short: "Cancel all queued jobs",
	Long:  "Cancel all queued jobs",
	Args:  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		ctx = mustAuthCtx(ctx)
		js, err := powClient.StorageJobs.Queued(ctx)
		checkErr(err)

		for _, j := range js.StorageJobs {
			_, err := powClient.StorageJobs.Cancel(ctx, j.Id)
			checkErr(err)
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
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		ctx = mustAuthCtx(ctx)
		js, err := powClient.StorageJobs.Executing(ctx)
		checkErr(err)

		for _, j := range js.StorageJobs {
			_, err := powClient.StorageJobs.Cancel(ctx, j.Id)
			checkErr(err)
		}
	},
}
