package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpaAddCmd.Flags().StringP("token", "t", "", "FPA access token")

	fpaCmd.AddCommand(fpaAddCmd)
}

var fpaAddCmd = &cobra.Command{
	Use:   "add [cid|path]",
	Short: "Add data to FPA via cid or file path",
	Long:  `Add data to FPA via a cid already in IPFS or local file path`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a valid cid or file path"))
		}

		token := viper.GetString("token")

		if token == "" {
			Fatal(errors.New("add requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpatoken"), token)

		var c *cid.Cid
		var f *os.File

		parsed, err := cid.Parse(args[0])
		if err == nil {
			c = &parsed
		} else {
			opened, err := os.Open(args[0])
			if err == nil {
				f = opened
				defer f.Close()
			}
		}

		if c != nil {
			s := spin.New("%s Adding specified cid to FPA...")
			s.Start()
			err = fcClient.Fpa.AddCid(ctx, *c)
			s.Stop()
			checkErr(err)
			Success("Added data for cid %s to FPA", c.String())
		} else if f != nil {
			s := spin.New("%s Adding specified file to FPA...")
			s.Start()
			cid, err := fcClient.Fpa.AddFile(ctx, f)
			s.Stop()
			checkErr(err)
			Success("Added file to FPA with resulting cid: %s", cid.String())
		} else {
			Fatal(errors.New("you must provide a valid cid or file path"))
		}
	},
}
