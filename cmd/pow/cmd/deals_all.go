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
)

func init() {
	dealsCmd.AddCommand(allCmd)
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "List pending and final deal records",
	Long:  `List pending and final deal records`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Retrieving deal records...")
		s.Start()
		res, err := fcClient.Deals.AllDealRecords(ctx)
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
		Message("Found %d deals", aurora.White(len(res)).Bold())
	},
}
