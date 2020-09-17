package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/ffs"
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

		job, err := fcClient.FFS.GetStorageJob(authCtx(ctx), jid)
		checkErr(err)
		dealErrorStrings := make([]string, len(job.DealErrors))
		for i, dealError := range job.DealErrors {
			dealErrorStrings[i] = dealError.Error()
		}
		rows := [][]string{
			{"API ID", job.APIID.String()},
			{"Job ID", job.ID.String()},
			{"CID", job.Cid.String()},
			{"Status", ffs.JobStatusStr[job.Status]},
			{"Error Cause", job.ErrCause},
			{"Deal Errors", strings.Join(dealErrorStrings, "\n")},
		}
		RenderTable(os.Stdout, []string{}, rows)
	},
}
