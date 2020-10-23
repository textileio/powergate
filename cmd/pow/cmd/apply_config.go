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
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	applyConfigCmd.Flags().StringP("conf", "c", "", "Optional path to a file containing storage config json, falls back to stdin, uses storage profile default by default")
	applyConfigCmd.Flags().BoolP("override", "o", false, "If set, override any pre-existing storage configuration for the cid")
	applyConfigCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	rootCmd.AddCommand(applyConfigCmd)
}

var applyConfigCmd = &cobra.Command{
	Use:   "apply-config [cid]",
	Short: "Apply the default or provided storage config to the specified cid",
	Long:  `Apply the default or provided storage config to the specified cid`,
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

		options := []client.ApplyStorageConfigOption{}

		if reader != nil {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			checkErr(err)

			config := &proto.StorageConfig{}
			err = protojson.UnmarshalOptions{}.Unmarshal(buf.Bytes(), config)
			checkErr(err)

			options = append(options, client.WithStorageConfig(config))
		}

		if viper.IsSet("override") {
			options = append(options, client.WithOverride(viper.GetBool("override")))
		}

		res, err := powClient.ApplyStorageConfig(mustAuthCtx(ctx), args[0], options...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))

		if viper.GetBool("watch") {
			watchJobIds(res.JobId)
		}
	},
}
