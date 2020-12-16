package apply

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	log = logging.Logger("apply")
)

func init() {
	Cmd.Flags().StringP("conf", "c", "", "Optional path to a file containing storage config json, falls back to stdin, uses the user default by default")
	Cmd.Flags().BoolP("override", "o", false, "If set, override any pre-existing storage configuration for the cid")
	Cmd.Flags().BoolP("noexec", "e", false, "If set, it doesn't create a job to ensure the new configuration")
	Cmd.Flags().BoolP("watch", "w", false, "Watch the progress of the resulting job")
	Cmd.Flags().StringSliceP("import-deals", "i", nil, "Comma-separated list of deal ids to import")
}

// Cmd is the command.
var Cmd = &cobra.Command{
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

		if viper.IsSet("noexec") {
			options = append(options, client.WithNoExec(viper.GetBool("noexec")))
		}

		if viper.IsSet("import-deals") {
			csvDealIDs := viper.GetStringSlice("import-deals")
			var dealIDs []uint64
			for _, strDealID := range csvDealIDs {
				dealID, err := strconv.ParseUint(strDealID, 10, 64)
				c.CheckErr(err)
				dealIDs = append(dealIDs, dealID)
			}
			options = append(options, client.WithImportDealIDs(dealIDs))
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
