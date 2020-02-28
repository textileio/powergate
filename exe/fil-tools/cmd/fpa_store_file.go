package cmd

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	storeFileCmd.Flags().StringP("file", "f", "", "Path to the file to store")
	storeFileCmd.Flags().StringP("token", "t", "", "FPA access token")

	fpaCmd.AddCommand(storeFileCmd)
}

var storeFileCmd = &cobra.Command{
	Use:   "store-file",
	Short: "Store file data in fpa",
	Long:  `Store file data in fpa`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		path := viper.GetString("file")
		token := viper.GetString("token")

		if token == "" {
			Fatal(errors.New("get requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpatoken"), token)

		file, err := os.Open(path)
		checkErr(err)
		defer file.Close()

		s := spin.New("%s Storing specified file in FPA...")
		s.Start()
		cid, err := fcClient.Fpa.StoreData(ctx, file)
		s.Stop()
		checkErr(err)

		Success("Stored data in FPA with resulting cid: %s", cid.String())
	},
}
