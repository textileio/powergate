package cmd

import (
	"context"
	"os"
	"strconv"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/textileio/filecoin/index/ask"
)

func init() {
	asksCmd.Flags().Uint64P("maxPrice", "m", 0, "max price for asks query")
	asksCmd.Flags().Uint64P("pieceSize", "p", 0, "piece size for asks query")
	asksCmd.Flags().IntP("limit", "l", -1, "limit the number of results")
	asksCmd.Flags().IntP("offset", "o", -1, "offset of results")

	rootCmd.AddCommand(asksCmd)
}

var asksCmd = &cobra.Command{
	Use:   "asks",
	Short: "Print the available asks",
	Long:  `Print the available asks`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

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

		asks, err := fcClient.Deals.AvailableAsks(ctx, q)
		checkErr(err)

		if len(asks) > 0 {
			data := make([][]string, len(asks))
			for i, a := range asks {
				data[i] = []string{
					a.Miner,
					strconv.Itoa(int(a.Price)),
					strconv.Itoa(int(a.MinPieceSize)),
					strconv.Itoa(int(a.Timestamp)),
					strconv.Itoa(int(a.Expiry)),
				}
			}
			RenderTable(os.Stdout, []string{"miner", "price", "min piece size", "timestamp", "expiry"}, data)
		}

		Message("Found %d asks", aurora.White(len(asks)).Bold())

	},
}
