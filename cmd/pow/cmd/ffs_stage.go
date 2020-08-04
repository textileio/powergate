package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/util"
)

func init() {
	ffsStageCmd.Flags().StringP("token", "t", "", "FFS access token")
	ffsStageCmd.Flags().String("ipfsrevproxy", "/ip4/127.0.0.1/tcp/6003", "Powergate IPFS reverse proxy multiaddr")

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
		if fi.IsDir() {
			cid, err = addFolder(args[0])
			if err != nil {
				Fatal(fmt.Errorf("staging folder: %s", err))
			}
		} else {
			f, err := os.Open(args[0])
			checkErr(err)
			defer func() { checkErr(f.Close()) }()

			s := spin.New("%s Staging specified file in FFS hot storage...")
			s.Start()
			ptrCid, err := fcClient.FFS.Stage(authCtx(ctx), f)
			s.Stop()
			checkErr(err)
			cid = *ptrCid
		}
		Success("Staged asset in FFS hot storage with cid: %s", util.CidToString(cid))
	},
}

func addFolder(folderPath string) (cid.Cid, error) {
	ipfsProxy := viper.GetString("ipfsrevproxy")
	ma, _ := multiaddr.NewMultiaddr(ipfsProxy)
	ipfs, err := httpapi.NewApi(ma)
	if err != nil {
		return cid.Undef, err
	}
	stat, err := os.Lstat(folderPath)
	if err != nil {
		return cid.Undef, err
	}
	ff, err := files.NewSerialFile(folderPath, false, stat)
	if err != nil {
		return cid.Undef, err
	}
	defer func() {
		if err := ff.Close(); err != nil {
			Fatal(fmt.Errorf("closing folder: %s", err))
		}
	}()
	opts := []options.UnixfsAddOption{
		options.Unixfs.CidVersion(1),
		options.Unixfs.Pin(false),
	}
	pth, err := ipfs.Unixfs().Add(context.Background(), files.ToDir(ff), opts...)
	if err != nil {
		return cid.Undef, err
	}
	return pth.Cid(), nil

}
