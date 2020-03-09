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
	retrieveCmd.Flags().StringP("address", "a", "", "wallet address to fund retrieval")
	retrieveCmd.Flags().StringP("cid", "c", "", "cid of the data to fetch")
	retrieveCmd.Flags().StringP("out", "o", "", "file path to write the data to")

	dealsCmd.AddCommand(retrieveCmd)
}

var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve data from filecoin",
	Long:  `Retrieve data from filecoin`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		addr := viper.GetString("address")
		cidString := viper.GetString("cid")
		out := viper.GetString("out")

		if addr == "" {
			Fatal(errors.New("get command needs a wallet address"))
		}

		if cidString == "" {
			Fatal(errors.New("get command needs a cid"))
		}

		if out == "" {
			Fatal(errors.New("get command needs an out path to write the data to"))
		}

		cid, err := cid.Parse(cidString)
		checkErr(err)

		s := spin.New("%s Retrieving specified data...")
		s.Start()
		reader, err := fcClient.Deals.Retrieve(ctx, addr, cid)
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
