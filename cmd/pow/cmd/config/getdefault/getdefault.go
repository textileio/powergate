package getdefault

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "default",
	Short: "Returns the default storage config",
	Long:  `Returns the default storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		res, err := c.PowClient.StorageConfig.Default(c.MustAuthCtx(ctx))
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.DefaultStorageConfig)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
