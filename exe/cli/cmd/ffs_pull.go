package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ffsPullCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsPullCmd)
}

var ffsPullCmd = &cobra.Command{
	Use:   "pull [cid]",
	Short: "Fetches the config for the provided cid",
	Long:  `Fetches the config for the provided cid`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide cid argument"))
		}

		c, err := cid.Parse(args[0])
		checkErr(err)

		s := spin.New("%s Pulling cid config...")
		s.Start()
		resp, err := fcClient.Ffs.GetCidConfig(authCtx(ctx), c)
		s.Stop()
		checkErr(err)

		json, err := json.MarshalIndent(resp.Config, "", "  ")
		checkErr(err)

		Message("Cid config:\n%s", string(json))
	},
}
