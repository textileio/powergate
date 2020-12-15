package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	storageJobsCmd.AddCommand(storageJobGetCmd)
}

var storageJobGetCmd = &cobra.Command{
	Use:   "get [jobid]",
	Short: "Get a storage job's current status",
	Long:  `Get a storage job's current status`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.StorageJobs.StorageJob(c.MustAuthCtx(ctx), args[0])
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.StorageJob)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
