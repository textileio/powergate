package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsSetDefaultCidConfigCmd.Flags().StringP("token", "t", "", "FFS auth token")
	ffsSetDefaultCidConfigCmd.Flags().StringP("file", "f", "", "optional path to file containing default cid config json, use stdin otherwise")

	ffsCmd.AddCommand(ffsSetDefaultCidConfigCmd)
}

var ffsSetDefaultCidConfigCmd = &cobra.Command{
	Use:   "setDefaultCidConfig",
	Short: "Sets the default cid storage config from stdin or a file",
	Long:  `Sets the default cid storage config from stdin or a file`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		path := viper.GetString("file")

		var reader io.Reader
		if len(path) > 0 {
			var err error
			reader, err = os.Open(path)
			checkErr(err)
		} else {
			reader = cmd.InOrStdin()
		}

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(reader)
		checkErr(err)

		config := ffs.DefaultCidConfig{}
		checkErr(json.Unmarshal(buf.Bytes(), &config))

		s := spin.New("%s Setting defautlt cid config...")
		s.Start()
		err = fcClient.Ffs.SetDefaultCidConfig(authCtx(ctx), config)
		s.Stop()
		checkErr(err)

		Success("Default cid storage config updated")
	},
}
