package stage

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
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().String("ipfsrevproxy", "127.0.0.1:6002", "Powergate IPFS reverse proxy multiaddr")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "stage [path|url]",
	Short: "Temporarily stage data in Hot Storage in preparation for applying a cid storage config",
	Long:  `Temporarily stage data in Hot Storage in preparation for applying a cid storage config`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()

		if strings.HasPrefix(strings.ToLower(args[0]), "http") {
			res, err := http.DefaultClient.Get(args[0])
			c.CheckErr(err)
			defer func() { c.CheckErr(res.Body.Close()) }()
			stageReader(ctx, res.Body)
			return
		}

		fi, err := os.Stat(args[0])
		if os.IsNotExist(err) {
			c.Fatal(errors.New("file/folder doesn't exist"))
		}
		if err != nil {
			c.Fatal(fmt.Errorf("getting file/folder information: %s", err))
		}
		if fi.IsDir() {
			cid, err := c.PowClient.Data.StageFolder(c.MustAuthCtx(ctx), viper.GetString("ipfsrevproxy"), args[0])
			c.CheckErr(err)
			c.Success("Staged folder with cid: %s", cid)
		} else {
			f, err := os.Open(args[0])
			c.CheckErr(err)
			defer func() { c.CheckErr(f.Close()) }()
			stageReader(ctx, f)
		}
	},
}

func stageReader(ctx context.Context, reader io.Reader) {
	res, err := c.PowClient.Data.Stage(c.MustAuthCtx(ctx), reader)
	c.CheckErr(err)

	json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
	c.CheckErr(err)

	fmt.Println(string(json))
}
