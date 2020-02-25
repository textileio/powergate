package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"path"

	"github.com/caarlos0/spin"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	fpaGetCmd.Flags().StringP("token", "t", "", "token of the request")
	fpaGetCmd.Flags().StringP("cid", "c", "", "cid of the data to fetch")
	fpaGetCmd.Flags().StringP("out", "o", "", "file path to write the data to")

	fpaCmd.AddCommand(fpaGetCmd)
}

var fpaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get data from fpa",
	Long:  `Get data from fpa`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		token := viper.GetString("token")
		cidString := viper.GetString("cid")
		out := viper.GetString("out")

		if token == "" {
			Fatal(errors.New("get requires token"))
		}
		ctx = context.WithValue(ctx, authKey("fpatoken"), token)

		if cidString == "" {
			Fatal(errors.New("store command needs a cid"))
		}

		if out == "" {
			Fatal(errors.New("get command needs an out path to write the data to"))
		}

		c, err := cid.Parse(cidString)
		checkErr(err)

		s := spin.New("%s Retrieving specified data...")
		s.Start()
		reader, err := fcClient.Fpa.Get(ctx, c)
		checkErr(err)

		dir := path.Dir(out)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			checkErr(err)
		}
		file, err := os.Create(out)
		checkErr(err)

		defer file.Close()

		buffer := make([]byte, 1024*32) // 32KB
		for {
			bytesRead, readErr := reader.Read(buffer)
			if readErr != nil && readErr != io.EOF {
				Fatal(readErr)
			}
			_, err = file.Write(buffer[:bytesRead])
			checkErr(err)
			if readErr == io.EOF {
				break
			}
		}
		s.Stop()
	},
}
