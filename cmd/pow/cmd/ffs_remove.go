package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsCmd.AddCommand(ffsRemoveCmd)
}

var ffsRemoveCmd = &cobra.Command{
	Use:   "remove [cid]",
	Short: "Removes a Cid from being tracked as an active storage",
	Long:  `Removes a Cid from being tracked as an active storage. The Cid should have both Hot and Cold storage disabled, if that isn't the case it will return ErrActiveInStorage`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		_, err := fcClient.FFS.Remove(mustAuthCtx(ctx), args[0])
		checkErr(err)
	},
}
