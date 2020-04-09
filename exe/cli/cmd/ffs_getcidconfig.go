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
	ffsGetCidConfigCmd.Flags().StringP("token", "t", "", "FFS auth token")
	ffsGetCidConfigCmd.Flags().StringP("cid", "c", "", "cid of the config to fetch")

	ffsCmd.AddCommand(ffsGetCidConfigCmd)
}

var ffsGetCidConfigCmd = &cobra.Command{
	Use:   "getCidConfig",
	Short: "Retuns the config for the provided cid",
	Long:  `Retuns the config for the provided cid`,
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
		resp, err := fcClient.Ffs.GetCidConfig(authCtx(ctx), c)
		s.Stop()
		checkErr(err)

		json, err := json.MarshalIndent(resp.Config, "", "  ")
		checkErr(err)

		Message("Cid config:\n%s", string(json))
	},
}
