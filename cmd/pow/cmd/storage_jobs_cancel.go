package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	storageJobsCmd.AddCommand(ffsCancelCmd)
}

var ffsCancelCmd = &cobra.Command{
	Use:   "cancel [jobid]",
	Short: "Cancel an executing storage job",
	Long:  `Cancel an executing storage job`,
	Args:  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := powClient.StorageJobs.CancelStorageJob(mustAuthCtx(ctx), args[0])
		checkErr(err)
	},
}
