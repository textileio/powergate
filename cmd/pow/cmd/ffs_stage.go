package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
)

func init() {
	ffsStageCmd.Flags().StringP("token", "t", "", "FFS access token")

	ffsCmd.AddCommand(ffsStageCmd)
}

var ffsStageCmd = &cobra.Command{
	Use:   "stage [path]",
	Short: "Temporarily cache data in the Hot layer in preparation for pushing a cid storage config",
	Long:  `Temporarily cache data in the Hot layer in preparation for pushing a cid storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a file path"))
		}

		f, err := os.Open(args[0])
		checkErr(err)
		defer func() { checkErr(f.Close()) }()

		s := spin.New("%s Caching specified file in FFS hot storage...")
		s.Start()
		cid, err := fcClient.FFS.Stage(authCtx(ctx), f)
		s.Stop()
		checkErr(err)
		Success("Cached file in FFS hot storage with cid: %s", util.CidToString(*cid))
	},
}
