package cmd

import (
	"context"
	"os"
	"strconv"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

func init() {
	asksCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the asks index",
	Long:  `Get the asks index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting storage asks...")
		s.Start()
		index, err := fcClient.Asks.Get(ctx)
		s.Stop()
		checkErr(err)

		if len(index.Storage) > 0 {
			Message("Storage median price price: %v", index.StorageMedianPrice)
			Message("Last updated: %v", index.LastUpdated.Format("01/02/06 15:04 MST"))
			data := make([][]string, len(index.Storage))
			i := 0
			for _, a := range index.Storage {
				data[i] = []string{
					a.Miner,
					strconv.Itoa(int(a.Price)),
					strconv.Itoa(int(a.MinPieceSize)),
					strconv.FormatInt(a.ChainEpoch, 10),
					strconv.FormatInt(a.Expiry, 10),
				}
				i++
			}
			RenderTable(os.Stdout, []string{"miner", "price", "min piece size", "chain epoch", "expiry"}, data)
		}

		Message("Found %d asks", aurora.White(len(index.Storage)).Bold())
	},
}
