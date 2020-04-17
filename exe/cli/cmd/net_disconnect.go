package cmd

import (
	"context"
	"errors"

	"github.com/caarlos0/spin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netDisconnectCmd)
}

var netDisconnectCmd = &cobra.Command{
	Use:   "disconnect [peerID] [address]",
	Short: "Disconnect from specified peer",
	Long:  `Disconnect from specified peer`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a peer id argument"))
		}

		peerID, err := peer.Decode(args[0])
		checkErr(err)

		s := spin.New("%s Disconnecting from peer...")
		s.Start()
		checkErr(err)
		err = fcClient.Net.DisconnectPeer(ctx, peerID)
		s.Stop()
		checkErr(err)

		Success("Disconnected from %v", peerID.String())
	},
}
