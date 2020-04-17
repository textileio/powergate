package cmd

import (
	"context"
	"encoding/json"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netPeersCmd)
}

var netPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Get the node peers",
	Long:  `Get the node peers`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting peers...")
		s.Start()
		peers, err := fcClient.Net.Peers(ctx)
		s.Stop()
		checkErr(err)

		bytes, err := json.MarshalIndent(peers, "", "  ")
		checkErr(err)

		Success(string(bytes))
	},
}
