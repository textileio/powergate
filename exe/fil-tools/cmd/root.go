package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	client "github.com/textileio/fil-tools/api/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	cfgFile string

	fcClient *client.Client

	cmdTimeout = time.Second * 10

	rootCmd = &cobra.Command{
		Use:   "filecoin",
		Short: "A client for storage and retreival of filecoin data",
		Long:  `A client for storage and retreival of filecoin data`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := viper.BindPFlag("serverAddress", cmd.Root().PersistentFlags().Lookup("serverAddress"))
			checkErr(err)

			target := viper.GetString("serverAddress")

			var creds credentials.TransportCredentials
			if strings.Contains(target, "443") {
				creds = credentials.NewTLS(&tls.Config{})
			}

			var opts []grpc.DialOption
			if creds != nil {
				opts = append(opts, grpc.WithTransportCredentials(creds))
			} else {
				opts = append(opts, grpc.WithInsecure())
			}

			auth := tokenAuth{}
			opts = append(opts, grpc.WithPerRPCCredentials(auth))
			fcClient, err = client.NewClient(target, opts...)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fil-tools.yaml)")
	rootCmd.PersistentFlags().String("serverAddress", "127.0.0.1:5002", "address of the filecoin service api")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		checkErr(err)

		// Search config in home directory with name ".fil-tools" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fil-tools")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

type authKey string

type tokenAuth struct {
	secure bool
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	md := map[string]string{}
	token, ok := ctx.Value(authKey("ffstoken")).(string)
	if ok && token != "" {
		md["X-ffs-Token"] = token
	}
	return md, nil
}

func (t tokenAuth) RequireTransportSecurity() bool {
	return t.secure
}
