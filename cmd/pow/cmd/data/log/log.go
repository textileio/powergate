package log

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	c "github.com/textileio/powergate/cmd/pow/common"
)

func init() {
	Cmd.Flags().StringP("jid", "j", "", "Display information for only this job id")
}

// Cmd displays logs for specified cid.
var Cmd = &cobra.Command{
	Use:     "log [cid]",
	Aliases: []string{"logs"},
	Short:   "Display logs for specified cid",
	Long:    `Display logs for specified cid`,
	Args:    cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
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

		err := c.PowClient.Data.WatchLogs(c.MustAuthCtx(ctx), ch, args[0], opts...)
		c.CheckErr(err)

		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-interrupt
			cancel()
			os.Exit(0)
		}()

		for {
			event, ok := <-ch
			if !ok {
				break
			}
			if event.Err != nil {
				c.Fatal(event.Err)
				break
			}
			ts := time.Unix(event.Res.LogEntry.Time, 0)
			c.Message("%v - %v", ts.Format("2006-01-02T15:04:05"), event.Res.LogEntry.Message)
		}
	},
}
