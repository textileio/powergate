package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs/rpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	ffsConfigPushCmd.Flags().StringP("conf", "c", "", "Optional path to a file containing storage config json, falls back to stdin, uses FFS default by default")
	ffsConfigPushCmd.Flags().BoolP("override", "o", false, "If set, override any pre-existing storage configuration for the cid")
	ffsConfigPushCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	ffsConfigCmd.AddCommand(ffsConfigPushCmd)
}

var ffsConfigPushCmd = &cobra.Command{
	Use:   "push [cid]",
	Short: "Add data to FFS via cid",
	Long:  `Add data to FFS via a cid already in IPFS`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

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

		options := []client.PushStorageConfigOption{}

		if reader != nil {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			checkErr(err)

			config := &rpc.StorageConfig{}
			err = protojson.UnmarshalOptions{}.Unmarshal(buf.Bytes(), config)
			checkErr(err)

			options = append(options, client.WithStorageConfig(config))
		}

		if viper.IsSet("override") {
			options = append(options, client.WithOverride(viper.GetBool("override")))
		}

		res, err := fcClient.FFS.PushStorageConfig(mustAuthCtx(ctx), args[0], options...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))

		if viper.GetBool("watch") {
			watchJobIds(res.JobId)
		}
	},
}
