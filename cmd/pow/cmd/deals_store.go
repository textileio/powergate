package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/deals"
)

func init() {
	storeCmd.Flags().StringP("cid", "c", "", "Data cid to store")
	storeCmd.Flags().Uint64P("size", "s", 0, "The data piece size")
	storeCmd.Flags().StringSliceP("miners", "m", []string{}, "miner ids of the deals to execute")
	storeCmd.Flags().IntSliceP("prices", "p", []int{}, "prices of the deals to execute")
	storeCmd.Flags().StringP("address", "a", "", "wallet address used to store the data")
	storeCmd.Flags().Uint64P("duration", "d", 0, "minimum duration to store the data for")

	dealsCmd.AddCommand(storeCmd)
}

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store data in filecoin",
	Long:  `Store data in filecoin`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		cidString := viper.GetString("cid")
		size := viper.GetUint64("size")
		addr := viper.GetString("address")
		minDuration := viper.GetUint64("duration")
		miners := viper.GetStringSlice("miners")
		prices := viper.GetIntSlice("prices")

		lMiners := len(miners)
		lPrices := len(prices)

		if cidString == "" {
			Fatal(errors.New("store command needs a data cid"))
		}

		if size == 0 {
			Fatal(errors.New("store command needs a size"))
		}

		if addr == "" {
			Fatal(errors.New("store command needs a wallet address"))
		}

		if minDuration == 0 {
			Fatal(errors.New("store command duration should be > 0"))
		}

		if lMiners == 0 {
			Fatal(errors.New("you must supply at least one miner"))
		}

		if lPrices == 0 {
			Fatal(errors.New("you must supply at least one price"))
		}

		if lMiners != lPrices {
			Fatal(fmt.Errorf("number of miners and prices must be equal but received %v miners and %v prices", lMiners, lPrices))
		}

		dealConfigs := make([]deals.StorageDealConfig, lMiners)
		for i, miner := range miners {
			dealConfigs[i] = deals.StorageDealConfig{
				Miner:      miner,
				EpochPrice: uint64(prices[i]),
			}
		}

		c, err := cid.Decode(cidString)
		checkErr(err)

		s := spin.New("%s Initiating specified storage deals...")
		s.Start()
		results, err := fcClient.Deals.Store(ctx, addr, c, size, dealConfigs, minDuration)
		s.Stop()
		checkErr(err)

		var failed []deals.StorageDealConfig
		var succeeded []cid.Cid
		for _, result := range results {
			if result.Success {
				succeeded = append(succeeded, result.ProposalCid)
			} else {
				failed = append(failed, result.Config)
			}
		}

		if len(succeeded) > 0 {
			Success("Initiation of %v deals succeeded with cids:", len(succeeded))
			for _, cid := range succeeded {
				Message(cid.String())
			}
		}
		if len(failed) > 0 {
			cmd.Println()
			Message("Failed to initialize %v deals:", len(failed))
			data := make([][]string, len(failed))
			for i, dealConfig := range failed {
				data[i] = []string{
					dealConfig.Miner,
					strconv.Itoa(int(dealConfig.EpochPrice)),
				}
			}
			RenderTable(os.Stdout, []string{"miner", "price"}, data)
		}

	},
}
