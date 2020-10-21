package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	netCmd.AddCommand(netFindCmd)
}

var netFindCmd = &cobra.Command{
	Use:   "find [peerID]",
	Short: "Find a peer by peer id",
	Long:  `Find a peer by peer id`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Net.FindPeer(ctx, args[0])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.PeerInfo)
		checkErr(err)

		fmt.Println(string(json))
	},
}
