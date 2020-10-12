package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
)

func init() {
	ffsStageCmd.Flags().String("ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")

	ffsCmd.AddCommand(ffsStageCmd)
}

var ffsStageCmd = &cobra.Command{
	Use:   "stage [path|url]",
	Short: "Temporarily stage data in the Hot layer in preparation for pushing a cid storage config",
	Long:  `Temporarily stage data in the Hot layer in preparation for pushing a cid storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a file/folder path"))
		}

		if strings.HasPrefix(strings.ToLower(args[0]), "http") {
			err := stageURL(ctx, args[0])
			checkErr(err)
			return
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
			cid, err = fcClient.FFS.StageFolder(mustAuthCtx(ctx), viper.GetString("ipfsrevproxy"), args[0])
			checkErr(err)
		} else {
			f, err := os.Open(args[0])
			checkErr(err)
			defer func() { checkErr(f.Close()) }()

			ptrCid, err := fcClient.FFS.Stage(mustAuthCtx(ctx), f)
			checkErr(err)
			cid = *ptrCid
		}
		s.Stop()
		Success("Staged asset in FFS hot storage with cid: %s", util.CidToString(cid))
	},
}

func stageURL(ctx context.Context, urlstr string) error {
	res, err := http.DefaultClient.Get(urlstr)
	if err != nil {
		return fmt.Errorf("GET %s: %w", urlstr, err)
	}

	defer func() { checkErr(res.Body.Close()) }()

	var cid cid.Cid
	s := spin.New("%s Staging URL in FFS hot storage...")
	s.Start()
	defer s.Stop()
	ptrCid, err := fcClient.FFS.Stage(authCtx(ctx), res.Body)
	if err != nil {
		return err
	}

	cid = *ptrCid
	s.Stop()
	Success("Staged asset in FFS hot storage with cid: %s", util.CidToString(cid))
	return nil
}
