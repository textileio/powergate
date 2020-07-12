package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	dealsCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status [proposal cid]",
	Short: "Returns the current status of the deal",
	Long:  `Returns the current status of the deal`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		c, err := cid.Decode(args[0])
		checkErr(err)

		s := spin.New("%s Getting deal status...")
		s.Start()
		code, slashed, err := fcClient.Deals.GetDealStatus(ctx, c)
		s.Stop()
		checkErr(err)

		Success("Deal status: %s", storagemarket.DealStates[code])
		if slashed {
			Message("Deal has been slashed!")
		}
	},
}
