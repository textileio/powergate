package cmd

import (
	"context"
	"errors"

	"github.com/caarlos0/spin"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
)

func init() {
	netCmd.AddCommand(netConnectCmd)
}

var netConnectCmd = &cobra.Command{
	Use:   "connect [peerID] [optional address]",
	Short: "Connect to a specified peer",
	Long:  `Connect to a specified peer`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a peer id argument"))
		}

		peerID, err := peer.Decode(args[0])
		checkErr(err)

		var addrs []ma.Multiaddr
		if len(args) == 2 {
			addr, err := ma.NewMultiaddr(args[1])
			checkErr(err)
			addrs = append(addrs, addr)
		}

		addrInfo := peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}

		s := spin.New("%s Connecting to peer...")
		s.Start()
		checkErr(err)
		err = fcClient.Net.ConnectPeer(ctx, addrInfo)
		s.Stop()
		checkErr(err)

		Success("Connected to %v", peerID.String())
	},
}
