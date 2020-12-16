package setdefault

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	log = logging.Logger("setdefault")
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "set-default [optional file]",
	Short: "Sets the default storage config from stdin or a file",
	Long:  `Sets the default storage config from stdin or a file`,
	Args:  cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		var reader io.Reader
		if len(args) > 0 {
			file, err := os.Open(args[0])
			defer func() {
				if err := file.Close(); err != nil {
					log.Errorf("closing config file: %s", err)
				}
			}()
			reader = file
			c.CheckErr(err)
		} else {
			reader = cmd.InOrStdin()
		}

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(reader)
		c.CheckErr(err)

		config := &userPb.StorageConfig{}
		err = protojson.UnmarshalOptions{}.Unmarshal(buf.Bytes(), config)
		c.CheckErr(err)

		_, err = c.PowClient.StorageConfig.SetDefault(c.MustAuthCtx(ctx), config)
		c.CheckErr(err)
	},
}
