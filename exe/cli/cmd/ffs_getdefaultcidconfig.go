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
	ffsGetDefaultCidConfigCmd.Flags().StringP("token", "t", "", "FFS auth token")
	ffsGetDefaultCidConfigCmd.Flags().StringP("cid", "c", "", "cid to use in the default config")

	ffsCmd.AddCommand(ffsGetDefaultCidConfigCmd)
}

var ffsGetDefaultCidConfigCmd = &cobra.Command{
	Use:   "getDefaultCidConfig",
	Short: "Retuns the default config prepped for the provided cid",
	Long:  `Retuns the default config prepped for the provided cid`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		cidString := viper.GetString("cid")
		if cidString == "" {
			Fatal(errors.New("cid flag required"))
		}

		c, err := cid.Parse(cidString)
		checkErr(err)

		s := spin.New("%s Getting defautlt cid config...")
		s.Start()
		resp, err := fcClient.Ffs.GetDefaultCidConfig(authCtx(ctx), c)
		s.Stop()
		checkErr(err)

		json, err := json.MarshalIndent(resp.Config, "", "  ")
		checkErr(err)

		Message("Default cid config:\n%s", string(json))
	},
}
