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
	fpaShowCmd.Flags().StringP("cid", "c", "", "cid of the data to pin")
	fpaShowCmd.Flags().StringP("token", "t", "", "wallet address used to store the data")

	fpaCmd.AddCommand(fpaShowCmd)
}

var fpaShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show pinned cid data",
	Long:  `Show pinned cid data`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		token := viper.GetString("token")
		cidString := viper.GetString("cid")

		if token == "" {
			Fatal(errors.New("get requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpatoken"), token)

		if cidString == "" {
			Fatal(errors.New("store command needs a cid"))
		}

		c, err := cid.Parse(cidString)
		checkErr(err)

		s := spin.New("%s Getting info from cid...")
		s.Start()
		info, err := fcClient.Fpa.Show(ctx, c)
		s.Stop()
		checkErr(err)

		buf, err := json.MarshalIndent(info, "", "  ")
		checkErr(err)
		Message("%s", buf)

	},
}
