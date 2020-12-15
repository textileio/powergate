package get

import (
	"context"
	"io"
	"os"
	"path"
	"time"

	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	c "github.com/textileio/powergate/cmd/pow/common"
)

func init() {
	Cmd.Flags().String("ipfsrevproxy", "localhost:6002", "Powergate IPFS reverse proxy DNS address. If port 443, is assumed is a HTTPS endpoint.")
	Cmd.Flags().BoolP("folder", "f", false, "Indicates that the retrieved Cid is a folder")
}

// Cmd get data stored by the user by cid.
var Cmd = &cobra.Command{
	Use:   "get [cid] [output file path]",
	Short: "Get data stored by the user by cid",
	Long:  `Get data stored by the user by cid`,
	Args:  cobra.ExactArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour*8)
		defer cancel()

		s := spin.New("%s Retrieving specified data...")
		s.Start()

		isFolder := viper.GetBool("folder")
		if isFolder {
			err := c.PowClient.Data.GetFolder(c.MustAuthCtx(ctx), viper.GetString("ipfsrevproxy"), args[0], args[1])
			c.CheckErr(err)
		} else {
			reader, err := c.PowClient.Data.Get(c.MustAuthCtx(ctx), args[0])
			c.CheckErr(err)

			dir := path.Dir(args[1])
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
				c.CheckErr(err)
			}
			file, err := os.Create(args[1])
			c.CheckErr(err)

			defer func() { c.CheckErr(file.Close()) }()

			buffer := make([]byte, 1024*32) // 32KB
			for {
				bytesRead, readErr := reader.Read(buffer)
				if readErr != nil && readErr != io.EOF {
					c.Fatal(readErr)
				}
				_, err = file.Write(buffer[:bytesRead])
				c.CheckErr(err)
				if readErr == io.EOF {
					break
				}
			}
		}
		s.Stop()
		c.Success("Data written to %v", args[1])
	},
}
