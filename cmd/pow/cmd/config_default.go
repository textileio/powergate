package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	configCmd.AddCommand(configDefaultCmd)
}

var configDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Returns the default storage config",
	Long:  `Returns the default storage config`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		res, err := powClient.StorageConfig.Default(mustAuthCtx(ctx))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.DefaultStorageConfig)
		checkErr(err)

		fmt.Println(string(json))
	},
}
