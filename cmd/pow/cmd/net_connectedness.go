package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	netCmd.AddCommand(netConnectednessCmd)
}

var netConnectednessCmd = &cobra.Command{
	Use:   "connectedness [peerID]",
	Short: "Check connectedness to a specified peer",
	Long:  `Check connectedness to a specified peer`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Net.Connectedness(ctx, args[0])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
