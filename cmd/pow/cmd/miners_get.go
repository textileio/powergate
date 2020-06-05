package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

func init() {
	minersCmd.AddCommand(getMinersCmd)
}

var getMinersCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the miners index",
	Long:  `Get the miners index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting miners data...")
		s.Start()
		index, err := fcClient.Miners.Get(ctx)
		s.Stop()
		checkErr(err)

		Message("%v", aurora.Blue("Miner metadata:").Bold())
		cmd.Println()

		Message("%v miners online", aurora.White(index.Meta.Online).Bold())
		Message("%v miners offline", aurora.White(index.Meta.Offline).Bold())
		cmd.Println()

		data := make([][]string, len(index.Meta.Info))
		i := 0
		for id, meta := range index.Meta.Info {
			data[i] = []string{
				id,
				meta.UserAgent,
				meta.Location.Country,
				fmt.Sprintf("%v", meta.Online),
				meta.LastUpdated.Format("01/02/06 15:04 MST"),
			}
			i++
		}
		RenderTable(os.Stdout, []string{"miner", "user agent", "location", "online", "last updated"}, data)

		Message("Found metadata for %d miners", aurora.White(len(index.Meta.Info)).Bold())
		cmd.Println()

		Message("%v", aurora.Blue("Miner on chain data:").Bold())
		cmd.Println()

		chainData := make([][]string, len(index.OnChain.Miners))
		i = 0
		for id, minerData := range index.OnChain.Miners {
			chainData[i] = []string{
				id,
				strconv.Itoa(int(minerData.Power)),
				strconv.Itoa(int(minerData.RelativePower)),
				strconv.Itoa(int(minerData.SectorSize)),
				strconv.Itoa(int(minerData.ActiveDeals)),
			}
			i++
		}

		RenderTable(os.Stdout, []string{"miner", "power", "relativePower", "sectorSize", "activeDeals"}, chainData)

		lastUpdated := time.Unix(index.OnChain.LastUpdated, 0).Format("01/02/06 15:04 MST")

		Message("Found on chain data for %d miners", aurora.White(len(index.OnChain.Miners)).Bold())
		Message("Chain data last updated %v", aurora.White(lastUpdated).Bold())
	},
}
