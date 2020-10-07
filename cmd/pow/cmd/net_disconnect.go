package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netDisconnectCmd)
}

var netDisconnectCmd = &cobra.Command{
	Use:   "disconnect [peerID]",
	Short: "Disconnect from specified peer",
	Long:  `Disconnect from specified peer`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		err := fcClient.Net.DisconnectPeer(ctx, args[0])
		checkErr(err)
	},
}
