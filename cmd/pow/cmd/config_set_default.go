package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	log = logging.Logger("cmd")
)

func init() {
	configCmd.AddCommand(configSetDefaultCmd)
}

var configSetDefaultCmd = &cobra.Command{
	Use:   "set-default [optional file]",
	Short: "Sets the default storage config from stdin or a file",
	Long:  `Sets the default storage config from stdin or a file`,
	Args:  cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
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
			checkErr(err)
		} else {
			reader = cmd.InOrStdin()
		}

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(reader)
		checkErr(err)

		config := &proto.StorageConfig{}
		err = protojson.UnmarshalOptions{}.Unmarshal(buf.Bytes(), config)
		checkErr(err)

		_, err = powClient.SetDefaultStorageConfig(mustAuthCtx(ctx), config)
		checkErr(err)
	},
}
