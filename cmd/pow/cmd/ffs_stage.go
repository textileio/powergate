package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
)

func init() {
	ffsStageCmd.Flags().StringP("token", "t", "", "FFS access token")
	ffsStageCmd.Flags().String("ipfsrevproxy", "127.0.0.1:6003", "Powergate IPFS reverse proxy multiaddr")

	ffsCmd.AddCommand(ffsStageCmd)
}

var ffsStageCmd = &cobra.Command{
	Use:   "stage [path]",
	Short: "Temporarily stage data in the Hot layer in preparation for pushing a cid storage config",
	Long:  `Temporarily stage data in the Hot layer in preparation for pushing a cid storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a file/folder path"))
		}

		fi, err := os.Stat(args[0])
		if os.IsNotExist(err) {
			Fatal(errors.New("file/folder doesn't exist"))
		}
		if err != nil {
			Fatal(fmt.Errorf("getting file/folder information: %s", err))
		}
		var cid cid.Cid
		s := spin.New("%s Staging specified asset in FFS hot storage...")
		s.Start()
		if fi.IsDir() {
			cid, err = fcClient.FFS.StageFolder(authCtx(ctx), viper.GetString("ipfsrevproxy"), args[0])
			checkErr(err)
		} else {
			f, err := os.Open(args[0])
			checkErr(err)
			defer func() { checkErr(f.Close()) }()

			ptrCid, err := fcClient.FFS.Stage(authCtx(ctx), f)
			checkErr(err)
			cid = *ptrCid
		}
		s.Stop()
		Success("Staged asset in FFS hot storage with cid: %s", util.CidToString(cid))
	},
}
