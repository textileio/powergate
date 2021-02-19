package retrievals

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client"
	c "github.com/textileio/powergate/v2/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	Cmd.Flags().BoolP("ascending", "a", false, "sort records ascending, default is descending")
	Cmd.Flags().StringSlice("cids", []string{}, "limit the records to deals for the specified data cids")
	Cmd.Flags().StringSlice("addrs", []string{}, "limit the records to deals initiated from  the specified wallet addresses")
	Cmd.Flags().BoolP("include-failed", "e", false, "include failed retrievals")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "retrievals",
	Short: "List retrieval deal records for the user",
	Long:  `List retrieval deal records for the user`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		var opts []client.DealRecordsOption

		if viper.IsSet("ascending") {
			opts = append(opts, client.WithAscending(viper.GetBool("ascending")))
		}
		if viper.IsSet("cids") {
			opts = append(opts, client.WithDataCids(viper.GetStringSlice("cids")...))
		}
		if viper.IsSet("addrs") {
			opts = append(opts, client.WithFromAddrs(viper.GetStringSlice("addrs")...))
		}
		if viper.IsSet("include-failed") {
			opts = append(opts, client.WithIncludeFailed(viper.GetBool("include-failed")))
		}

		res, err := c.PowClient.Deals.RetrievalDealRecords(c.MustAuthCtx(ctx), opts...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
