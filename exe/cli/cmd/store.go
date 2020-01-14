package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/lotus/types"
)

var success []cid.Cid
var failed []deals.DealConfig

func init() {
	storeCmd.Flags().StringP("file", "f", "", "Path to the file to store")
	storeCmd.Flags().StringSliceP("miners", "m", []string{}, "miner ids of the deals to execute")
	storeCmd.Flags().Int64SliceP("prices", "p", []int64{}, "prices of the deals to execute")
	storeCmd.Flags().Uint64P("duration", "d", 0, "the duration to store the file for")

	rootCmd.AddCommand(storeCmd)

	success = make([]cid.Cid, 5)
	for i := range success {
		cid, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPU")
		success[i] = cid
	}
	failed = []deals.DealConfig{
		deals.DealConfig{
			Miner:      "asdf3rerf34",
			EpochPrice: types.NewInt(4567),
		},
		deals.DealConfig{
			Miner:      "kjn34kjnf",
			EpochPrice: types.NewInt(2345),
		},
		deals.DealConfig{
			Miner:      "ubsdc7askj434",
			EpochPrice: types.NewInt(8576),
		},
	}
}

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store data in filecoin",
	Long:  `Store data in filecoin`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		addr, err := cmd.Flags().GetString("address")
		checkErr(err)

		if addr == "" {
			Fatal(errors.New("balance command needs a wallet address"))
		}

		path, err := cmd.Flags().GetString("file")
		checkErr(err)

		file, err := os.Open(path)
		checkErr(err)

		miners, err := cmd.Flags().GetStringSlice("miners")
		checkErr(err)

		prices, err := cmd.Flags().GetInt64Slice("prices")
		checkErr(err)

		lMiners := len(miners)
		lPrices := len(prices)

		if lMiners == 0 {
			Fatal(errors.New("you must supply at least one miner"))
		}

		if lPrices == 0 {
			Fatal(errors.New("you must supply at least one price"))
		}

		if lMiners != lPrices {
			Fatal(fmt.Errorf("number of miners and prices must be equal but received %v miners and %v prices", lMiners, lPrices))
		}

		dealConfigs := make([]deals.DealConfig, lMiners)
		for i, miner := range miners {
			dealConfigs[i] = deals.DealConfig{
				Miner:      miner,
				EpochPrice: types.NewInt(uint64(prices[i])),
			}
		}

		duration, err := cmd.Flags().GetUint64("duration")
		checkErr(err)

		cmd.Println(ctx, addr, file, dealConfigs, duration)
		// success, failed, err = fcClient.Deals.Store(ctx, addr, file, dealConfigs, duration)
		// checkErr(err)

		if len(success) > 0 {
			Success("Initiation of %v deals succeeded with cids:", len(success))
			for _, cid := range success {
				Message(cid.String())
			}
		}
		if len(failed) > 0 {
			cmd.Println()
			Message("Failed to initialize %v deals:", len(success))
			data := make([][]string, len(failed))
			for i, dealConfig := range failed {
				data[i] = []string{
					dealConfig.Miner,
					strconv.Itoa(int(dealConfig.EpochPrice.Int64())),
				}
			}
			RenderTable([]string{"miner", "price"}, data)
		}

	},
}
