package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/deals"
)

func init() {
	storeCmd.Flags().StringP("file", "f", "", "Path to the file to store")
	storeCmd.Flags().StringSliceP("miners", "m", []string{}, "miner ids of the deals to execute")
	storeCmd.Flags().IntSliceP("prices", "p", []int{}, "prices of the deals to execute")
	storeCmd.Flags().StringP("address", "a", "", "wallet address used to store the data")
	storeCmd.Flags().Uint64P("duration", "d", 0, "duration to store the data for")

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

		addr := viper.GetString("address")
		duration := viper.GetUint64("duration")
		path := viper.GetString("file")
		miners := viper.GetStringSlice("miners")
		prices := viper.GetIntSlice("prices")

		lMiners := len(miners)
		lPrices := len(prices)

		if addr == "" {
			Fatal(errors.New("store command needs a wallet address"))
		}

		if duration == 0 {
			Fatal(errors.New("store command duration should be > 0"))
		}

		file, err := os.Open(path)
		checkErr(err)
		defer checkErr(file.Close())

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

		s := spin.New("%s Initiating specified storage deals...")
		s.Start()
		success, failed, err := fcClient.Deals.Store(ctx, addr, file, dealConfigs, duration)
		s.Stop()
		checkErr(err)

		if len(success) > 0 {
			Success("Initiation of %v deals succeeded with cids:", len(success))
			for _, cid := range success {
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
