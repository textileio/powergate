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
	client "github.com/textileio/filecoin/api/stub"
	"google.golang.org/grpc/credentials"
)

var (
	// Used for flags.
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.filecoin.yaml)")
	rootCmd.PersistentFlags().StringP("address", "a", "", "the default wallet address")
	rootCmd.PersistentFlags().Uint64P("duration", "d", 0, "the duration to store the file for")
	rootCmd.PersistentFlags().Uint64("maxPrice", 0, "max price for asks query")
	rootCmd.PersistentFlags().Uint64("pieceSize", 0, "piece size for asks query")
	rootCmd.PersistentFlags().StringSlice("wallets", []string{}, "existing wallets to switch between")

	err := viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))
	checkErr(err)
	err = viper.BindPFlag("duration", rootCmd.PersistentFlags().Lookup("duration"))
	checkErr(err)
	err = viper.BindPFlag("maxPrice", rootCmd.PersistentFlags().Lookup("maxPrice"))
	checkErr(err)
	err = viper.BindPFlag("pieceSize", rootCmd.PersistentFlags().Lookup("pieceSize"))
	checkErr(err)
	err = viper.BindPFlag("wallets", rootCmd.PersistentFlags().Lookup("wallets"))
	checkErr(err)

	viper.SetDefault("address", "")
	viper.SetDefault("duration", 100)
	viper.SetDefault("maxPrice", 200)
	viper.SetDefault("pieceSize", 1024)
	viper.SetDefault("wallets", []string{})
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		checkErr(err)

		// Search config in home directory with name ".filecoin" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".filecoin")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
