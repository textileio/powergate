package cmd

import (
	"context"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	dealsCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch [address] [cid]",
	Short: "Fetches deal data to the underlying blockstore of the Filecoin client",
	Long:  `Fetches deal data to the underlying blockstore of the Filecoin client.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		c, err := cid.Decode(args[1])
		checkErr(err)

		s := spin.New("%s Initiating fetch...")
		s.Start()
		err = fcClient.Deals.Fetch(ctx, args[0], c)
		s.Stop()
		checkErr(err)

		Success("Initiated fetch for data cid %v", c.String())
	},
}
