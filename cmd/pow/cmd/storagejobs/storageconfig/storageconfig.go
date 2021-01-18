package storageconfig

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "storage-config [job-id]",
	Short: "Get the StorageConfig associated with the specified job",
	Long:  `Get the StorageConfig associated with the specified job`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.StorageJobs.StorageConfigForJob(c.MustAuthCtx(ctx), args[0])
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.StorageConfig)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
