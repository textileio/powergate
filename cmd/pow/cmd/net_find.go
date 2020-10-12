package cmd

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/caarlos0/spin"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netFindCmd)
}

var netFindCmd = &cobra.Command{
	Use:   "find [peerID]",
	Short: "Find a peer by peer id",
	Long:  `Find a peer by peer id`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a peer id argument"))
		}

		s := spin.New("%s Finding peer...")
		s.Start()
		peerID, err := peer.Decode(args[0])
		checkErr(err)
		peerInfo, err := fcClient.Net.FindPeer(ctx, peerID)
		s.Stop()
		checkErr(err)

		bytes, err := json.MarshalIndent(peerInfo, "", "  ")
		checkErr(err)

		Success(string(bytes))
	},
}
