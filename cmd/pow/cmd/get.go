package cmd

import (
	"context"
	"io"
	"os"
	"path"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	getCmd.Flags().String("ipfsrevproxy", "localhost:6002", "Powergate IPFS reverse proxy DNS address. If port 443, is assumed is a HTTPS endpoint.")
	getCmd.Flags().BoolP("folder", "f", false, "Indicates that the retrieved Cid is a folder")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [cid] [output file path]",
	Short: "Get data by cid from the storage profile",
	Long:  `Get data by cid from the storage profile`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()

		s := spin.New("%s Retrieving specified data...")
		s.Start()

		isFolder := viper.GetBool("folder")
		if isFolder {
			err := powClient.GetFolder(mustAuthCtx(ctx), viper.GetString("ipfsrevproxy"), args[0], args[1])
			checkErr(err)
		} else {
			reader, err := powClient.Get(mustAuthCtx(ctx), args[0])
			checkErr(err)

			dir := path.Dir(args[1])
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
				checkErr(err)
			}
			file, err := os.Create(args[1])
			checkErr(err)

			defer func() { checkErr(file.Close()) }()

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
		}
		s.Stop()
		Success("Data written to %v", args[1])
	},
}
