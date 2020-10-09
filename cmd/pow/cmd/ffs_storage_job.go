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
	ffsCmd.AddCommand(ffsStorageJobsCmd)
	ffsStorageJobsCmd.AddCommand(
		ffsGetStorageJobCmd,
		ffsQueuedStorageJobsCmd,
		ffsExecutingStorageJobsCmd,
		ffsLatestFinalStorageJobsCmd,
		ffsLatestSuccessfulStorageJobsCmd,
	)
}

var ffsStorageJobsCmd = &cobra.Command{
	Use:     "storage-jobs",
	Aliases: []string{"storage-job"},
	Short:   "Privdes commands to query for storage jobs in various states",
	Long:    `Privdes commands to query for storage jobs in various statess`,
}

var ffsGetStorageJobCmd = &cobra.Command{
	Use:     "get [jobid]",
	Aliases: []string{"storage-job"},
	Short:   "Get a storage job's current status",
	Long:    `Get a storage job's current status`,
	Args:    cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
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

var ffsQueuedStorageJobsCmd = &cobra.Command{
	Use:   "queued [optional cid1,cid2,...]",
	Short: "List queued storage jobs",
	Long:  `List queued storage jobs`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.GetQueuedStorageJobs(authCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var ffsExecutingStorageJobsCmd = &cobra.Command{
	Use:   "executing [optional cid1,cid2,...]",
	Short: "List executing storage jobs",
	Long:  `List executing storage jobs`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.GetExecutingStorageJobs(authCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var ffsLatestFinalStorageJobsCmd = &cobra.Command{
	Use:   "latest-final [optional cid1,cid2,...]",
	Short: "List the latest final storage jobs",
	Long:  `List the latest final storage jobs`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.GetLatestFinalStorageJobs(authCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var ffsLatestSuccessfulStorageJobsCmd = &cobra.Command{
	Use:   "latest-successful [optional cid1,cid2,...]",
	Short: "List the latest successful storage jobs",
	Long:  `List the latest successful storage jobs`,
	Args:  cobra.RangeArgs(0, 1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cids []string
		if len(args) > 0 {
			cids = strings.Split(args[0], ",")
		}

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.FFS.GetLatestSuccessfulStorageJobs(authCtx(ctx), cids...)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
