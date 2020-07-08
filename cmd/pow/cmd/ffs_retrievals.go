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
	ffsrRetrievalsCmd.Flags().BoolP("ascending", "a", false, "sort records ascending, default is descending")
	ffsrRetrievalsCmd.Flags().StringSlice("cids", []string{}, "limit the records to deals for the specified data cids")
	ffsrRetrievalsCmd.Flags().StringSlice("addrs", []string{}, "limit the records to deals initiated from  the specified wallet addresses")

	ffsCmd.AddCommand(ffsrRetrievalsCmd)
}

var ffsrRetrievalsCmd = &cobra.Command{
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

		s := spin.New("%s Getting retrieval records...")
		s.Start()
		res, err := fcClient.FFS.ListRetrievalDealRecords(authCtx(ctx), opts...)
		s.Stop()
		checkErr(err)

		if len(res) > 0 {
			data := make([][]string, len(res))
			for i, r := range res {
				t := time.Unix(r.Time, 0)
				data[i] = []string{
					t.Format("01/02/06 15:04 MST"),
					r.Addr,
					r.DealInfo.Miner,
					r.DealInfo.RootCid.String(),
					strconv.Itoa(int(r.DealInfo.Size)),
				}
			}
			RenderTable(os.Stdout, []string{"time", "addr", "miner", "piece cid", "size"}, data)
		}
		Message("Found %d retrieval deal records", aurora.White(len(res)).Bold())
	},
}
