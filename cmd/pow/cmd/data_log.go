package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
)

func init() {
	dataLogCmd.Flags().StringP("jid", "j", "", "Display information for only this job id")

	dataCmd.AddCommand(dataLogCmd)
}

var dataLogCmd = &cobra.Command{
	Use:     "log [cid]",
	Aliases: []string{"logs"},
	Short:   "Display logs for specified cid",
	Long:    `Display logs for specified cid`,
	Args:    cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		opts := []client.WatchLogsOption{client.WithHistory(true)}
		jid := viper.GetString("jid")
		if jid != "" {
			opts = append(opts, client.WithJobIDFilter(jid))
		}

		ch := make(chan client.WatchLogsEvent)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := powClient.Data.WatchLogs(mustAuthCtx(ctx), ch, args[0], opts...)
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
			ts := time.Unix(event.Res.LogEntry.Time, 0)
			Message("%v - %v", ts.Format("2006-01-02T15:04:05"), event.Res.LogEntry.Message)
		}
	},
}
