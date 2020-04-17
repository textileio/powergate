package cmd

import (
	"context"
	"os"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(healthCmd)
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Display the node health status",
	Long:  `Display the node health status`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		s := spin.New("%s Checking node health...")
		s.Start()
		status, messages, err := fcClient.Health.Check(ctx)
		s.Stop()
		checkErr(err)

		Success("Health status: %v", status.String())
		if len(messages) > 0 {
			rows := make([][]string, len(messages))
			for i, message := range messages {
				rows[i] = []string{message}
			}
			RenderTable(os.Stdout, []string{"messages"}, rows)
		}
	},
}
