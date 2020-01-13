package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/filecoin/api/client"
	"google.golang.org/grpc/credentials"
)

var (
	// Used for flags.
	address string
	cfgFile string

	fcClient *client.Client

	cmdTimeout = time.Second * 10

	rootCmd = &cobra.Command{
		Use:   "filecoin",
		Short: "A client for storage and retreival of filecoin data",
		Long:  `A client for storage and retreival of filecoin data`,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			// ToDo: get this fom config
			const target = "127.0.0.1:5002"
			var creds credentials.TransportCredentials
			if strings.Contains(target, "443") {
				creds = credentials.NewTLS(&tls.Config{})
			}
			var err error
			fcClient, err = client.NewClient(target, creds)
			checkErr(err)
		},
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringP("address", "a", "", "wallet address to run command on")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("address", "")

	// rootCmd.AddCommand(addCmd)
	// rootCmd.AddCommand(initCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		checkErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
