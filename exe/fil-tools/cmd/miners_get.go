package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

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
				meta.LastUpdated.String(),
			}
			i++
		}
		RenderTable(os.Stdout, []string{"miner", "user agent", "location", "online", "last updated"}, data)

		Message("Found metadata for %d miners", aurora.White(len(index.Meta.Info)).Bold())
		cmd.Println()

		Message("%v", aurora.Blue("Miner on chain data:").Bold())
		cmd.Println()

		chainData := make([][]string, len(index.Chain.Power))
		i = 0
		for id, power := range index.Chain.Power {
			chainData[i] = []string{
				id,
				strconv.Itoa(int(power.Power)),
				strconv.Itoa(int(power.Relative)),
			}
			i++
		}

		RenderTable(os.Stdout, []string{"miner", "power", "relative"}, chainData)

		Message("Found on chain data for %d miners", aurora.White(len(index.Chain.Power)).Bold())
		Message("Chain data last updated %v", aurora.White(index.Chain.LastUpdated).Bold())
	},
}
