package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	dataStageCmd.Flags().String("ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")

	dataCmd.AddCommand(dataStageCmd)
}

var dataStageCmd = &cobra.Command{
	Use:   "stage [path|url]",
	Short: "Temporarily stage data in the Hot layer in preparation for applying a cid storage config",
	Long:  `Temporarily stage data in the Hot layer in preparation for applying a cid storage config`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()

		if strings.HasPrefix(strings.ToLower(args[0]), "http") {
			res, err := http.DefaultClient.Get(args[0])
			checkErr(err)
			defer func() { checkErr(res.Body.Close()) }()
			stageReader(ctx, res.Body)
			return
		}

		fi, err := os.Stat(args[0])
		if os.IsNotExist(err) {
			Fatal(errors.New("file/folder doesn't exist"))
		}
		if err != nil {
			Fatal(fmt.Errorf("getting file/folder information: %s", err))
		}
		if fi.IsDir() {
			c, err := powClient.Data.StageFolder(mustAuthCtx(ctx), viper.GetString("ipfsrevproxy"), args[0])
			checkErr(err)
			Success("Staged folder with cid: %s", c)
		} else {
			f, err := os.Open(args[0])
			checkErr(err)
			defer func() { checkErr(f.Close()) }()
			stageReader(ctx, f)
		}
	},
}

func stageReader(ctx context.Context, reader io.Reader) {
	res, err := powClient.Data.Stage(mustAuthCtx(ctx), reader)
	checkErr(err)

	json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
	checkErr(err)

	fmt.Println(string(json))
}
