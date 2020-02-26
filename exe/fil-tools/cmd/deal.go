package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/fil-tools/deals"
	"github.com/textileio/fil-tools/index/ask"
)

func init() {
	dealCmd.Flags().StringP("address", "a", "", "wallet address used to store the data")
	dealCmd.Flags().Uint64P("duration", "d", 0, "duration to store the data for")
	dealCmd.Flags().StringP("file", "f", "", "path to the file to store")
	dealCmd.Flags().Uint64P("maxPrice", "m", 0, "max price of the asks to query")
	dealCmd.Flags().IntP("pieceSize", "p", 0, "piece size of the asks to query")
	dealCmd.Flags().IntP("limit", "l", -1, "limit the number of asks results")
	dealCmd.Flags().IntP("offset", "o", -1, "offset of asks results")

	rootCmd.AddCommand(dealCmd)
}

var dealCmd = &cobra.Command{
	Use:   "deal",
	Short: "Interactive storage deal",
	Long:  `Interactive storage deal`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		asksCtx, asksCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer asksCancel()

		addr := viper.GetString("address")
		duration := viper.GetUint64("duration")
		path := viper.GetString("file")
		mp := viper.GetUint64("maxPrice")
		ps := viper.GetUint64("pieceSize")
		l := viper.GetInt("limit")
		o := viper.GetInt("offset")

		if addr == "" {
			Fatal(errors.New("deal command needs a wallet address"))
		}

		if duration == 0 {
			Fatal(errors.New("deal command duration should be > 0"))
		}

		if mp == 0 {
			Fatal(errors.New("maxPrice must be > 0"))
		}

		if ps == 0 {
			Fatal(errors.New("pieceSize must be > 0"))
		}

		file, err := os.Open(path)
		checkErr(err)

		q := ask.Query{
			MaxPrice:  mp,
			PieceSize: ps,
			Limit:     l,
			Offset:    o,
		}

		s := spin.New("%s Querying network for available storage asks...")
		s.Start()
		asks, err := fcClient.Asks.Query(asksCtx, q)
		s.Stop()
		checkErr(err)

		if len(asks) == 0 {
			Message("No available asks, exiting")
			return
		}

		choices := make([]string, len(asks))
		for i, ask := range asks {
			expiry := time.Unix(int64(ask.Expiry), 0).Format("01/02/06 15:04 MST")
			choices[i] = fmt.Sprintf("Price of %v with min piece size %v expiring %v", ask.Price, ask.MinPieceSize, expiry)
		}

		selectedAsks := []int{}
		prompt := &survey.MultiSelect{
			Message: "Select from available asks:",
			Options: choices,
		}
		survey.AskOne(prompt, &selectedAsks, survey.WithValidator(survey.Required))
		cmd.Println()

		dealConfigs := make([]deals.StorageDealConfig, len(selectedAsks))
		for i, val := range selectedAsks {
			ask := asks[val]
			dealConfigs[i] = deals.StorageDealConfig{
				Miner:      ask.Miner,
				EpochPrice: ask.Price,
			}
		}

		storeCtx, storeCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer storeCancel()

		s = spin.New("%s Initiating selected storage deals...")
		s.Start()
		succeeded, failed, err := fcClient.Deals.Store(storeCtx, addr, file, dealConfigs, duration)
		s.Stop()
		checkErr(err)

		if len(failed) > 0 {
			Message("Failed to initialize %v deals:", len(failed))
			cmd.Println()
			data := make([][]string, len(failed))
			for i, dealConfig := range failed {
				data[i] = []string{
					dealConfig.Miner,
					strconv.Itoa(int(dealConfig.EpochPrice)),
				}
			}
			RenderTable(os.Stdout, []string{"miner", "price"}, data)
			cmd.Println()
			time.Sleep(time.Second * 2)
		}

		if len(succeeded) == 0 {
			Message("Failed to initialize any deals, exiting")
			return
		}

		Message("Watching progress of %v deals:", len(succeeded))
		cmd.Println()

		watchCtx, watchCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer watchCancel()

		processWatchEventsUntilDone(watchCtx, succeeded)
	},
}
