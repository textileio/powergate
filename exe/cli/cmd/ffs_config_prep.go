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
	ffsConfigPrepCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsConfigCmd.AddCommand(ffsConfigPrepCmd)
}

var ffsConfigPrepCmd = &cobra.Command{
	Use:   "prep [cid]",
	Short: "Returns the default config prepped for the provided cid",
	Long:  `Returns the default config prepped for the provided cid`,
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

		s := spin.New("%s Getting default cid config...")
		s.Start()
		config, err := fcClient.FFS.GetDefaultCidConfig(authCtx(ctx), c)
		s.Stop()
		checkErr(err)

		json, err := json.MarshalIndent(config, "", "  ")
		checkErr(err)

		Message("Default cid config:\n%s", string(json))
	},
}
