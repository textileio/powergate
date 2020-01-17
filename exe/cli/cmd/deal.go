package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gosuri/uilive"
	"github.com/spf13/cobra"
	"github.com/textileio/filecoin/api/client"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/index/ask"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/caarlos0/spin"
)

func init() {
	dealCmd.Flags().StringP("file", "f", "", "Path to the file to store")
	dealCmd.Flags().Uint64P("duration", "d", 0, "the duration to store the file for")
	dealCmd.Flags().Uint64P("maxPrice", "m", 0, "max price for asks query")
	dealCmd.Flags().Uint64P("pieceSize", "p", 0, "piece size for asks query")
	dealCmd.Flags().IntP("limit", "l", -1, "limit the number of results")
	dealCmd.Flags().IntP("offset", "o", -1, "offset of results")

	rootCmd.AddCommand(dealCmd)
}

var dealCmd = &cobra.Command{
	Use:   "deal",
	Short: "Interactive storage deal",
	Long:  `Interactive storage deal`,
	Run: func(cmd *cobra.Command, args []string) {
		asksCtx, asksCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer asksCancel()

		addr, err := cmd.Flags().GetString("address")
		checkErr(err)

		if addr == "" {
			Fatal(errors.New("deal command needs a wallet address"))
		}

		path, err := cmd.Flags().GetString("file")
		checkErr(err)
		file, err := os.Open(path)
		checkErr(err)
		duration, err := cmd.Flags().GetUint64("duration")
		checkErr(err)
		mp, err := cmd.Flags().GetUint64("maxPrice")
		checkErr(err)
		ps, err := cmd.Flags().GetUint64("pieceSize")
		checkErr(err)
		l, err := cmd.Flags().GetInt("limit")
		checkErr(err)
		o, err := cmd.Flags().GetInt("offset")
		checkErr(err)

		q := ask.Query{
			MaxPrice:  mp,
			PieceSize: ps,
			Limit:     l,
			Offset:    o,
		}

		s := spin.New("%s Querying network for available storage asks...")
		s.Start()
		asks, err := fcClient.Deals.AvailableAsks(asksCtx, q)
		s.Stop()
		checkErr(err)

		if len(asks) == 0 {
			Message("No available asks, exiting")
			return
		}

		choices := make([]string, len(asks))
		for i, ask := range asks {
			choices[i] = fmt.Sprintf("Price of %v with min piece size %v expiring %v", ask.Price, ask.MinPieceSize, ask.Expiry)
		}

		selectedAsks := []int{}
		prompt := &survey.MultiSelect{
			Message: "Select from available asks:",
			Options: choices,
		}
		survey.AskOne(prompt, &selectedAsks, survey.WithValidator(survey.Required))
		cmd.Println()

		dealConfigs := make([]deals.DealConfig, len(selectedAsks))
		for i, val := range selectedAsks {
			ask := asks[val]
			dealConfigs[i] = deals.DealConfig{
				Miner:      ask.Miner,
				EpochPrice: types.NewInt(ask.Price),
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
					strconv.Itoa(int(dealConfig.EpochPrice.Int64())),
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

		state := make(map[string]*client.WatchEvent, len(succeeded))
		for _, cid := range succeeded {
			state[cid.String()] = nil
		}

		writer := uilive.New()
		writer.Start()

		updateOutput(writer, state)

		watchCtx, watchCancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer watchCancel()

		ch, err := fcClient.Deals.Watch(watchCtx, succeeded)
		checkErr(err)

		for {
			event, ok := <-ch
			if ok == false {
				break
			}
			state[event.Deal.ProposalCid.String()] = &event
			updateOutput(writer, state)
			if isComplete(state) {
				break
			}
		}
		writer.Stop()
	},
}
