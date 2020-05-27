package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsAddToHotCmd.Flags().StringP("token", "t", "", "FFS access token")

	ffsCmd.AddCommand(ffsAddToHotCmd)
}

var ffsAddToHotCmd = &cobra.Command{
	Use:   "addToHot [path]",
	Short: "Add data to FFS hot storage via file path",
	Long:  `Add data to FFS hot storage via file path`,
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

		s := spin.New("%s Adding specified file to FFS hot storage...")
		s.Start()
		cid, err := fcClient.FFS.AddToHot(authCtx(ctx), f)
		s.Stop()
		checkErr(err)
		Success("Added file to FFS hot storage with cid: %s", cid.String())
	},
}
