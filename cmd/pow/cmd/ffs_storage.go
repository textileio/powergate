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
	ffsStorageCmd.Flags().BoolP("ascending", "a", false, "sort records ascending, default is sort descending")
	ffsStorageCmd.Flags().StringSlice("cids", []string{}, "limit the records to deals for the specified data cids, treated as and AND operation if --addrs is also provided")
	ffsStorageCmd.Flags().StringSlice("addrs", []string{}, "limit the records to deals initiated from  the specified wallet addresses, treated as and AND operation if --cids is also provided")
	ffsStorageCmd.Flags().BoolP("include-pending", "p", false, "include pending deals")
	ffsStorageCmd.Flags().BoolP("include-final", "f", false, "include final deals")

	ffsCmd.AddCommand(ffsStorageCmd)
}

var ffsStorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "List storage deal records for an FFS instance",
	Long:  `List storage deal records for an FFS instance`,
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
		if viper.IsSet("include-pending") {
			opts = append(opts, client.WithIncludePending(viper.GetBool("include-pending")))
		}
		if viper.IsSet("include-final") {
			opts = append(opts, client.WithIncludeFinal(viper.GetBool("include-final")))
		}

		res, err := fcClient.FFS.ListStorageDealRecords(mustAuthCtx(ctx), opts...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
