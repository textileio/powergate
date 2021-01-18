package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().StringSlice("user-ids", nil, "filter results by provided user ids.")
	Cmd.Flags().StringSlice("cids", nil, "filter results by provided cids.")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "Returns a list of information about all stored cids, filtered by user ids and cids if provided.",
	Long:  `Returns a list of information about all stored cids, filtered by user ids and cids if provided.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		userIDs := viper.GetStringSlice("user-ids")
		cids := viper.GetStringSlice("cids")

		res, err := c.PowClient.Admin.StorageInfo.ListStorageInfo(c.MustAuthCtx(ctx), userIDs, cids)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
