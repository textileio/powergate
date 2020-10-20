package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	netCmd.AddCommand(netListenAddrCmd)
}

var netListenAddrCmd = &cobra.Command{
	Use:   "addr",
	Short: "Get the listen address of the node",
	Long:  `Get the listen address of the node`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Net.ListenAddr(ctx)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.AddrInfo)
		checkErr(err)

		fmt.Println(string(json))
	},
}
