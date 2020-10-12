package cmd

import (
	"context"
	"encoding/json"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
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

		s := spin.New("%s Getting listen address...")
		s.Start()
		addrInfo, err := fcClient.Net.ListenAddr(ctx)
		s.Stop()
		checkErr(err)

		bytes, err := json.MarshalIndent(addrInfo, "", "  ")
		checkErr(err)

		Success(string(bytes))
	},
}
