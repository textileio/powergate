package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
)

func init() {
	ffsAddrsNewCmd.Flags().StringP("format", "f", "", "Optionally specify address format bls or secp256k1")
	ffsAddrsNewCmd.Flags().BoolP("default", "d", false, "Make the new address the ffs default")

	ffsAddrsCmd.AddCommand(ffsAddrsNewCmd)
}

var ffsAddrsNewCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new wallet address",
	Long:  `Create a new wallet address`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("must provide a name for the address"))
		}

		format := viper.GetString("format")
		makeDefault := viper.GetBool("default")

		var opts []client.NewAddressOption
		if format != "" {
			opts = append(opts, client.WithAddressType(format))
		}
		if makeDefault {
			opts = append(opts, client.WithMakeDefault(makeDefault))
		}

		s := spin.New("%s Getting FFS instance wallet address...")
		s.Start()
		addr, err := fcClient.FFS.NewAddr(authCtx(ctx), args[0], opts...)
		s.Stop()
		checkErr(err)
		Success("Created new wallet address: %s", addr)
	},
}
