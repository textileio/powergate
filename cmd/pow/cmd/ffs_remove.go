package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
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
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must a cid arguments"))
		}

		c, err := cid.Parse(args[0])
		checkErr(err)

		s := spin.New("%s Removing cid config...")
		s.Start()
		err = fcClient.FFS.Remove(mustAuthCtx(ctx), c)
		s.Stop()
		checkErr(err)
		Success("Removed cid config")
	},
}
