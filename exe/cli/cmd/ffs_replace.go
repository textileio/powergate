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
	ffsReplaceCmd.Flags().StringP("token", "t", "", "FFS access token")
	ffsReplaceCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	ffsCmd.AddCommand(ffsReplaceCmd)
}

var ffsReplaceCmd = &cobra.Command{
	Use:   "replace [cid1] [cid2]",
	Short: "Pushes a CidConfig of c2 equal to c1, and removes c1",
	Long:  `Pushes a CidConfig of c2 equal to c1, and removes c1. This operation is more efficient than manually removing and adding in two separate operations`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 2 {
			Fatal(errors.New("you must provide two cid arguments"))
		}

		c1, err := cid.Parse(args[0])
		checkErr(err)
		c2, err := cid.Parse(args[1])
		checkErr(err)

		s := spin.New("%s Replacing cid configuration...")
		s.Start()
		jid, err := fcClient.Ffs.Replace(authCtx(ctx), c1, c2)
		s.Stop()
		checkErr(err)
		Success("Replaced cid config with job id: %v", jid.String())

		if viper.GetBool("watch") {
			watchJobIds(jid)
		}
	},
}
