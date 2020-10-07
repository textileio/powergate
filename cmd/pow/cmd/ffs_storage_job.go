package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	ffsStorageJobCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsStorageJobCmd)
}

var ffsStorageJobCmd = &cobra.Command{
	Use:   "storage-job [jobid]",
	Short: "Get a storage job's current status",
	Long:  `Get a storage job's current status`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		idStrings := strings.Split(args[0], ",")
		jobIds := make([]ffs.JobID, len(idStrings))
		for i, s := range idStrings {
			jobIds[i] = ffs.JobID(s)
		}

		jid := ffs.JobID(args[0])

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.GetStorageJob(authCtx(ctx), jid)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res.Job)
		checkErr(err)

		fmt.Println(string(json))
	},
}
