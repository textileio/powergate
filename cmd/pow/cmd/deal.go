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
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/index/ask"
)

func init() {
	dealCmd.Flags().StringP("address", "a", "", "wallet address used to store the data")
	dealCmd.Flags().Uint64P("duration", "d", 0, "duration to store the data for")
	dealCmd.Flags().StringP("file", "f", "", "path to the file to store")
	dealCmd.Flags().BoolP("car", "c", false, "specifies if the file is already CAR encoded")
	dealCmd.Flags().Uint64P("max-price", "m", 0, "max price of the asks to query")
	dealCmd.Flags().IntP("size", "s", 0, "piece size of the asks to query")
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
		minDuration := viper.GetUint64("min-duration")
		path := viper.GetString("file")
		isCar := viper.GetBool("car")
		mp := viper.GetUint64("max-price")
		ps := viper.GetUint64("size")
		l := viper.GetInt("limit")
		o := viper.GetInt("offset")

		if addr == "" {
			Fatal(errors.New("deal command needs a wallet address"))
		}

		if minDuration == 0 {
			Fatal(errors.New("deal command minDuration should be > 0"))
		}

		if mp == 0 {
			Fatal(errors.New("max-price must be > 0"))
		}

		if ps == 0 {
			Fatal(errors.New("piece-size must be > 0"))
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
			expiry := time.Unix(ask.Expiry, 0).Format("01/02/06 15:04 MST")
			choices[i] = fmt.Sprintf("Price of %v with min piece size %v expiring %v", ask.Price, ask.MinPieceSize, expiry)
		}

		selectedAsks := []int{}
		prompt := &survey.MultiSelect{
			Message: "Select from available asks:",
			Options: choices,
		}
		checkErr(survey.AskOne(prompt, &selectedAsks, survey.WithValidator(survey.Required)))
		cmd.Println()

		dealConfigs := make([]deals.StorageDealConfig, len(selectedAsks))
		for i, val := range selectedAsks {
			ask := asks[val]
			dealConfigs[i] = deals.StorageDealConfig{
				Miner:      ask.Miner,
				EpochPrice: ask.Price,
			}
		}

		importCtx, importCancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer importCancel()

		s = spin.New("%s Importing data into the Filecoin client...")
		s.Start()

		c, size, err := fcClient.Deals.Import(importCtx, file, isCar)
		s.Stop()
		checkErr(err)
		Message("Imported %v bytes", size)

		s = spin.New("%s Initiating selected storage deals...")
		s.Start()
		storeCtx, storeCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer storeCancel()
		results, err := fcClient.Deals.Store(storeCtx, addr, c, uint64(size), dealConfigs, minDuration)
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
