package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/textileio/powergate/net/rpc"
)

func init() {
	netCmd.AddCommand(netConnectCmd)
}

var netConnectCmd = &cobra.Command{
	Use:   "connect [peerID] [optional address]",
	Short: "Connect to a specified peer",
	Long:  `Connect to a specified peer`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		if len(args) != 1 {
			Fatal(errors.New("you must provide a peer id argument"))
		}

		var addrs []string
		if len(args) == 2 {
			addrs = append(addrs, args[1])
		}

		addrInfo := &rpc.PeerAddrInfo{
			Id:    args[0],
			Addrs: addrs,
		}

		err := fcClient.Net.ConnectPeer(ctx, addrInfo)
		checkErr(err)
	},
}
