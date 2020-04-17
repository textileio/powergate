package cmd

import (
	"context"
	"errors"

	"github.com/caarlos0/spin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netConnectednessCmd)
}

var netConnectednessCmd = &cobra.Command{
	Use:   "connectedness [peerID]",
	Short: "Check connectedness to a specified peer",
	Long:  `Check connectedness to a specified peer`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a peer id argument"))
		}

		peerID, err := peer.Decode(args[0])
		checkErr(err)

		s := spin.New("%s Checking connectedness to peer...")
		s.Start()
		checkErr(err)
		connectedness, err := fcClient.Net.Connectedness(ctx, peerID)
		s.Stop()
		checkErr(err)

		Success("Connectedness: %v", connectedness.String())
	},
}
