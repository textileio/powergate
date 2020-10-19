package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsLogCmd.Flags().StringP("jid", "j", "", "Display information for only this job id")

	ffsCmd.AddCommand(ffsLogCmd)
}

var ffsLogCmd = &cobra.Command{
	Use:   "log [cid]",
	Short: "Display logs for specified cid",
	Long:  `Display logs for specified cid`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			Fatal(errors.New("you must provide a cid"))
		}

		cid, err := cid.Parse(args[0])
		checkErr(err)

		opts := []client.WatchLogsOption{client.WithHistory(true)}
		jid := viper.GetString("jid")
		if jid != "" {
			opts = append(opts, client.WithJidFilter(ffs.JobID(jid)))
		}

		ch := make(chan client.WatchLogsEvent)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = fcClient.FFS.WatchLogs(mustAuthCtx(ctx), ch, cid, opts...)
		checkErr(err)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cancel()
			os.Exit(0)
		}()

		for {
			event, ok := <-ch
			if !ok {
				break
			}
			if event.Err != nil {
				Fatal(event.Err)
				break
			}
			Message("%v - %v", event.LogEntry.Timestamp.Format("2006-01-02T15:04:05"), event.LogEntry.Msg)
		}
	},
}
