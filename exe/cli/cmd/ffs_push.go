package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsPushCmd.Flags().StringP("token", "t", "", "FFS access token")
	ffsPushCmd.Flags().StringP("config", "c", "", "Optional path to a file containing cid storage config json, falls back to stdin, uses FFS default by default")
	ffsPushCmd.Flags().BoolP("override", "o", false, "Path to a file containing cid storage config json")

	ffsCmd.AddCommand(ffsPushCmd)
}

var ffsPushCmd = &cobra.Command{
	Use:   "push [cid]",
	Short: "Add data to FFS via cid",
	Long:  `Add data to FFS via a cid already in IPFS`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a cid"))
		}

		c, err := cid.Parse(args[0])
		checkErr(err)

		configPath := viper.GetString("config")

		var reader io.Reader
		if len(configPath) > 0 {
			file, err := os.Open(configPath)
			defer file.Close()
			reader = file
			checkErr(err)
		} else {
			stat, _ := os.Stdin.Stat()
			// stdin is being piped in (not being read from terminal)
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				reader = cmd.InOrStdin()
			}
		}

		options := []client.PushConfigOption{}

		if reader != nil {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			checkErr(err)

			config := ffs.CidConfig{Cid: c}
			checkErr(json.Unmarshal(buf.Bytes(), &config))

			options = append(options, client.WithCidConfig(config))
		}

		if viper.IsSet("override") {
			options = append(options, client.WithOverride(viper.GetBool("override")))
		}

		s := spin.New("%s Adding cid storage config to FFS...")
		s.Start()
		jid, err := fcClient.Ffs.PushConfig(authCtx(ctx), c, options...)
		s.Stop()
		checkErr(err)
		Success("Pushed cid config for %s to FFS with job id: %v", c.String(), jid.String())
	},
}
