package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/caarlos0/spin"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logging.Logger("cmd")
)

func init() {
	ffsConfigSetCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsConfigCmd.AddCommand(ffsConfigSetCmd)
}

var ffsConfigSetCmd = &cobra.Command{
	Use:   "set [(optional)file]",
	Short: "Sets the default cid storage config from stdin or a file",
	Long:  `Sets the default cid storage config from stdin or a file`,
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

		config := ffs.StorageConfig{}
		checkErr(json.Unmarshal(buf.Bytes(), &config))

		s := spin.New("%s Setting default storage config...")
		s.Start()
		err = fcClient.FFS.SetDefaultConfig(authCtx(ctx), config)
		s.Stop()
		checkErr(err)

		Success("Default storage config updated")
	},
}
