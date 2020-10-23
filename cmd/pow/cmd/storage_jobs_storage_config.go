package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	storageJobsCmd.AddCommand(storageJobsStorageConfigCmd)
}

var storageJobsStorageConfigCmd = &cobra.Command{
	Use:   "storage-config [job-id]",
	Short: "Get the StorageConfig associated with the specified job",
	Long:  `Get the StorageConfig associated with the specified job`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := powClient.StorageJobs.StorageConfigForJob(mustAuthCtx(ctx), args[0])
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res.StorageConfig)
		checkErr(err)

		fmt.Println(string(json))
	},
}
