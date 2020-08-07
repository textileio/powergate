package cmd

import (
	"context"
	"encoding/json"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(serverInfoCmd)
}

var serverInfoCmd = &cobra.Command{
	Use:   "server-info",
	Short: "Display information about the connected server",
	Long:  `Display information about the connected server`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Getting server info...")
		s.Start()
		info, err := fcClient.BuildInfo(ctx)
		s.Stop()
		checkErr(err)

		bytes, err := json.MarshalIndent(info, "", "    ")
		checkErr(err)

		Success("\nHost: %v\nBuild info:\n%v", fcClient.Host(), string(bytes))
	},
}
