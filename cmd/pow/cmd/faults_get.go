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
	faultsCmd.AddCommand(getFaultsCmd)
}

var getFaultsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the faults index",
	Long:  `Get the faults index`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting Faults data...")
		s.Start()
		index, err := fcClient.Faults.Get(ctx)
		s.Stop()
		checkErr(err)

		Message("%v", aurora.Blue("Faults index data:").Bold())
		cmd.Println()

		Message("Tipset key: %v", aurora.White(index.TipSetKey).Bold())
		cmd.Println()

		data := make([][]string, len(index.Miners))
		i := 0
		for id, faults := range index.Miners {
			data[i] = []string{
				id,
				fmt.Sprintf("%v", faults.Epochs),
			}
			i++
		}
		RenderTable(os.Stdout, []string{"miner", "faults"}, data)

		Message("Found faults data for %d miners", aurora.White(len(index.Miners)).Bold())
	},
}
