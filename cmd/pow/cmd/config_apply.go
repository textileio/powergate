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
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	configApplyCmd.Flags().StringP("conf", "c", "", "Optional path to a file containing storage config json, falls back to stdin, uses the user default by default")
	configApplyCmd.Flags().BoolP("override", "o", false, "If set, override any pre-existing storage configuration for the cid")
	configApplyCmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")

	configCmd.AddCommand(configApplyCmd)
}

var configApplyCmd = &cobra.Command{
	Use:   "apply [cid]",
	Short: "Apply the default or provided storage config to the specified cid",
	Long:  `Apply the default or provided storage config to the specified cid`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
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
			c.CheckErr(err)
		} else {
			stat, _ := os.Stdin.Stat()
			// stdin is being piped in (not being read from terminal)
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				reader = cmd.InOrStdin()
			}
		}

		options := []client.ApplyOption{}

		if reader != nil {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(reader)
			c.CheckErr(err)

			config := &userPb.StorageConfig{}
			err = protojson.UnmarshalOptions{}.Unmarshal(buf.Bytes(), config)
			c.CheckErr(err)

			options = append(options, client.WithStorageConfig(config))
		}

		if viper.IsSet("override") {
			options = append(options, client.WithOverride(viper.GetBool("override")))
		}

		res, err := c.PowClient.StorageConfig.Apply(c.MustAuthCtx(ctx), args[0], options...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))

		if viper.GetBool("watch") {
			c.WatchJobIds(res.JobId)
		}
	},
}
