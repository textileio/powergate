package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsCmd.AddCommand(ffsCancelCmd)
}

var ffsCancelCmd = &cobra.Command{
	Use:   "cancel [jobid]",
	Short: "Cancel an executing job",
	Long:  `Cancel an executing job`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			Fatal(errors.New("you must provide a job id"))
		}

		jid := ffs.JobID(args[0])
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		s := spin.New("%s Signaling executing Job cancellation...")
		s.Start()
		err := fcClient.FFS.CancelJob(mustAuthCtx(ctx), jid)
		s.Stop()
		checkErr(err)
		Success("Successful cancellation signaling.")

	},
}
