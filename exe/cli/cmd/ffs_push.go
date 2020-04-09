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
	ffsAddCmd.Flags().StringP("token", "t", "", "FFS access token")

	ffsCmd.AddCommand(ffsAddCmd)
}

var ffsAddCmd = &cobra.Command{
	Use:   "push [cid]",
	Short: "Add data to FFS via cid",
	Long:  `Add data to FFS via a cid already in IPFS`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a cid"))
		}

		token := viper.GetString("token")

		if token == "" {
			Fatal(errors.New("add requires token"))
		}
		ctx = context.WithValue(ctx, authKey("ffstoken"), token)

		c, err := cid.Parse(args[0])
		checkErr(err)

		s := spin.New("%s Adding specified cid to FFS...")
		s.Start()
		jid, err := fcClient.Ffs.PushConfig(ctx, c)
		s.Stop()
		checkErr(err)
		Success("Added data for cid %s to FFS with job id: %v", c.String(), jid.String())
	},
}
