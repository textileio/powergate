package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCmd.AddCommand(ffsCancelCmd)
}

var ffsCancelCmd = &cobra.Command{
	Use:   "cancel [jobid]",
	Short: "Cancel an executing job",
	Long:  `Cancel an executing job`,
	Args:  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := fcClient.FFS.CancelJob(mustAuthCtx(ctx), args[0])
		checkErr(err)
	},
}
