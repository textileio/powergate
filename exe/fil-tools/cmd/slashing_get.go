package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/spin"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

func init() {
	slashingCmd.AddCommand(getSlashingCmd)
}

var getSlashingCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the slashing index",
	Long:  `Get the slashing index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting slashing data...")
		s.Start()
		index, err := fcClient.Slashing.Get(ctx)
		s.Stop()
		checkErr(err)

		Message("%v", aurora.Blue("Slashing index data:").Bold())
		cmd.Println()

		Message("Tipset key: %v", aurora.White(index.TipSetKey).Bold())
		cmd.Println()

		data := make([][]string, len(index.Miners))
		i := 0
		for id, slashes := range index.Miners {
			data[i] = []string{
				id,
				fmt.Sprintf("%v", slashes.Epochs),
			}
			i++
		}
		RenderTable(os.Stdout, []string{"miner", "slashes"}, data)

		Message("Found slashes data for %d miners", aurora.White(len(index.Miners)).Bold())
	},
}
