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
	ffsConfigPushCmd.Flags().StringP("token", "t", "", "FFS access token")
	ffsConfigPushCmd.Flags().StringP("conf", "c", "", "Optional path to a file containing storage config json, falls back to stdin, uses FFS default by default")
	ffsConfigPushCmd.Flags().BoolP("override", "o", false, "If set, override any pre-existing storage configuration for the cid")
	ffsConfigPushCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	ffsConfigCmd.AddCommand(ffsConfigPushCmd)
}

var ffsConfigPushCmd = &cobra.Command{
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

		configPath := viper.GetString("conf")

		var reader io.Reader
		if len(configPath) > 0 {
			file, err := os.Open(configPath)
			defer func() {
				if err := file.Close(); err != nil {
					log.Errorf("closing config file: %s", err)
				}
			}()
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

			config := ffs.StorageConfig{}
			checkErr(json.Unmarshal(buf.Bytes(), &config))

			options = append(options, client.WithStorageConfig(config))
		}

		if viper.IsSet("override") {
			options = append(options, client.WithOverride(viper.GetBool("override")))
		}

		s := spin.New("%s Adding cid storage config to FFS...")
		s.Start()
		jid, err := fcClient.FFS.PushConfig(authCtx(ctx), c, options...)
		s.Stop()
		checkErr(err)
		Success("Pushed cid storage config for %s to FFS with job id: %v", c.String(), jid.String())

		if viper.GetBool("watch") {
			watchJobIds(jid)
		}
	},
}
