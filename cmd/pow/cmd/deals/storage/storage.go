package storage

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
	Cmd.Flags().BoolP("ascending", "a", false, "sort records ascending, default is sort descending")
	Cmd.Flags().StringSlice("cids", []string{}, "limit the records to deals for the specified data cids, treated as and AND operation if --addrs is also provided")
	Cmd.Flags().StringSlice("addrs", []string{}, "limit the records to deals initiated from  the specified wallet addresses, treated as and AND operation if --cids is also provided")
	Cmd.Flags().BoolP("include-pending", "p", false, "include pending deals")
	Cmd.Flags().BoolP("include-final", "f", false, "include final deals")
	Cmd.Flags().BoolP("include-failed", "e", false, "include failed deals")
}

// Cmd is the command.
var Cmd = &cobra.Command{
	Use:   "storage",
	Short: "List storage deal records for the user",
	Long:  `List storage deal records for the user`,
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
		if viper.IsSet("include-pending") {
			opts = append(opts, client.WithIncludePending(viper.GetBool("include-pending")))
		}
		if viper.IsSet("include-final") {
			opts = append(opts, client.WithIncludeFinal(viper.GetBool("include-final")))
		}
		if viper.IsSet("include-failed") {
			opts = append(opts, client.WithIncludeFailed(viper.GetBool("include-failed")))
		}

		res, err := c.PowClient.Deals.StorageDealRecords(c.MustAuthCtx(ctx), opts...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}
