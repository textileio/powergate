package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	ffsRetrievalsCmd.Flags().BoolP("ascending", "a", false, "sort records ascending, default is descending")
	ffsRetrievalsCmd.Flags().StringSlice("cids", []string{}, "limit the records to deals for the specified data cids")
	ffsRetrievalsCmd.Flags().StringSlice("addrs", []string{}, "limit the records to deals initiated from  the specified wallet addresses")

	ffsCmd.AddCommand(ffsRetrievalsCmd)
}

var ffsRetrievalsCmd = &cobra.Command{
	Use:   "retrievals",
	Short: "List retrieval deal records for an FFS instance",
	Long:  `List retrieval deal records for an FFS instance`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		var opts []client.ListDealRecordsOption

		if viper.IsSet("ascending") {
			opts = append(opts, client.WithAscending(viper.GetBool("ascending")))
		}
		if viper.IsSet("cids") {
			opts = append(opts, client.WithDataCids(viper.GetStringSlice("cids")...))
		}
		if viper.IsSet("addrs") {
			opts = append(opts, client.WithFromAddrs(viper.GetStringSlice("addrs")...))
		}

		res, err := fcClient.FFS.ListRetrievalDealRecords(mustAuthCtx(ctx), opts...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
