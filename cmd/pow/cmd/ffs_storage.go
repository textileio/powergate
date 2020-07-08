package cmd

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
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

		s := spin.New("%s Retrieving deal records...")
		s.Start()
		res, err := fcClient.FFS.ListStorageDealRecords(authCtx(ctx), opts...)
		s.Stop()
		checkErr(err)

		if len(res) > 0 {
			data := make([][]string, len(res))
			for i, r := range res {
				t := time.Unix(r.Time, 0)
				pending := ""
				if r.Pending {
					pending = "pending"
				}
				data[i] = []string{
					pending,
					strconv.FormatInt(r.DealInfo.ActivationEpoch, 10),
					t.Format("01/02/06 15:04 MST"),
					r.Addr,
					r.DealInfo.Miner,
					strconv.Itoa(int(r.DealInfo.DealID)),
					strconv.Itoa(int(r.DealInfo.PricePerEpoch)),
					r.DealInfo.PieceCID.String(),
					strconv.Itoa(int(r.DealInfo.Size)),
					strconv.Itoa(int(r.DealInfo.Duration)),
				}
			}
			RenderTable(os.Stdout, []string{"pending", "active epoch", "time", "addr", "miner", "deal id", "price/epoch", "piece cid", "size", "duration"}, data)
		}
		Message("Found %d storage deal records", aurora.White(len(res)).Bold())
	},
}
