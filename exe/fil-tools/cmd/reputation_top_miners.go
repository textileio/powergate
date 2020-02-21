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
	topMinersCmd.Flags().IntP("limit", "l", -1, "limit the number of results")

	reputationCmd.AddCommand(topMinersCmd)
}

var topMinersCmd = &cobra.Command{
	Use:   "topMiners",
	Short: "Fetches a list of the currently top rated miners",
	Long:  `Fetches a list of the currently top rated miners`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		limit, err := cmd.Flags().GetInt("limit")
		checkErr(err)

		s := spin.New("%s Fetching top miners...")
		s.Start()
		topMiners, err := fcClient.Reputation.GetTopMiners(ctx, limit)
		s.Stop()
		checkErr(err)

		data := make([][]string, len(topMiners))
		for i, minerScore := range topMiners {
			data[i] = []string{
				minerScore.Addr,
				strconv.Itoa(minerScore.Score),
			}
		}

		RenderTable(os.Stdout, []string{"miner", "score"}, data)

		Message("Showing data for %d miners", aurora.White(len(topMiners)).Bold())
	},
}
