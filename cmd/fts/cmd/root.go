package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	client "github.com/textileio/powergate/api/client"
)

var (
	powClient *client.Client

	cmdTimeout = time.Second * 10

	rootCmd = &cobra.Command{
		Use:               "fts",
		Short:             "Filecoin data transfer service client for migrating big data",
		Long:              `Filecoin data transfer service client for migrating big data`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlag("serverAddress", cmd.Root().PersistentFlags().Lookup("serverAddress"))
			checkErr(err)

			target := viper.GetString("serverAddress")

			powClient, err = client.NewClient(target)
			checkErr(err)
		},
		Run: func(cmd *cobra.Command, args []string) {
			version, err := cmd.Flags().GetBool("version")
			checkErr(err)
			if version {
				versionCmd.Run(cmd, args)
			} else {
				err := cmd.Help()
				checkErr(err)
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
	rootCmd.Flags().BoolP("version", "v", false, "display version information for fts and the connected server")
	rootCmd.PersistentFlags().String("serverAddress", "127.0.0.1:5002", "address of the powergate service api")
	rootCmd.PersistentFlags().StringP("token", "t", "", "user auth token")
}

func initConfig() {
	viper.SetEnvPrefix("POW")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}
