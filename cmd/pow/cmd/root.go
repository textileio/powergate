package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	client "github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/cmd/pow/cmd/data"
	c "github.com/textileio/powergate/cmd/pow/common"
)

var (
	rootCmd = &cobra.Command{
		Use:               "pow",
		Short:             "A client for storage and retreival of powergate data",
		Long:              `A client for storage and retreival of powergate data`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlag("serverAddress", cmd.Root().PersistentFlags().Lookup("serverAddress"))
			c.CheckErr(err)

			target := viper.GetString("serverAddress")

			c.PowClient, err = client.NewClient(target)
			c.CheckErr(err)
		},
		Run: func(cmd *cobra.Command, args []string) {
			version, err := cmd.Flags().GetBool("version")
			c.CheckErr(err)
			if version {
				versionCmd.Run(cmd, args)
			} else {
				err := cmd.Help()
				c.CheckErr(err)
			}
		},
	}
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("version", "v", false, "display version information for pow and the connected server")
	rootCmd.PersistentFlags().String("serverAddress", "127.0.0.1:5002", "address of the powergate service api")
	rootCmd.PersistentFlags().StringP("token", "t", "", "user auth token")

	rootCmd.AddCommand(data.Cmd)
}

func initConfig() {
	viper.SetEnvPrefix("POW")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}
