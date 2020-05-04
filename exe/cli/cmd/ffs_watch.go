package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/gosuri/uilive"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/ffs"
)

func init() {
	ffsWatchCmd.Flags().StringP("token", "t", "", "FFS auth token")

	ffsCmd.AddCommand(ffsWatchCmd)
}

var ffsWatchCmd = &cobra.Command{
	Use:   "watch [jobid,...]",
	Short: "Watch for job status updates",
	Long:  `Watch for job status updates`,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			Fatal(errors.New("you must provide a comma-separated list of job ids"))
		}

		idStrings := strings.Split(args[0], ",")
		jobIds := make([]ffs.JobID, len(idStrings))
		for i, s := range idStrings {
			jobIds[i] = ffs.JobID(s)
		}

		watchJobIds(jobIds...)
	},
}

func watchJobIds(jobIds ...ffs.JobID) {
	state := make(map[string]*client.JobEvent, len(jobIds))
	for _, jobID := range jobIds {
		state[jobID.String()] = nil
	}

	writer := uilive.New()
	writer.Start()

	updateJobsOutput(writer, state)

	ch := make(chan client.JobEvent)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := fcClient.Ffs.WatchJobs(authCtx(ctx), ch, jobIds...)
	checkErr(err)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		os.Exit(0)
	}()

	for {
		event, ok := <-ch
		if !ok {
			break
		}
		state[event.Job.ID.String()] = &event
		updateJobsOutput(writer, state)
		if jobsComplete(state) {
			break
		}
	}

	writer.Stop()
}

func updateJobsOutput(writer io.Writer, state map[string]*client.JobEvent) {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := make([][]string, len(keys))
	for i, k := range keys {
		val := "awaiting state"
		if state[k] != nil {
			if state[k].Job.Status == ffs.Failed {
				val = fmt.Sprintf("%v %v", displayName(state[k].Job.Status), state[k].Job.ErrCause)
			} else if state[k].Err != nil {
				val = fmt.Sprintf("Error: %v", state[k].Err.Error())
			} else {
				val = displayName(state[k].Job.Status)
			}
		}
		data[i] = []string{
			k,
			val,
		}
	}
	RenderTable(writer, []string{"Job id", "Status"}, data)
}

func jobsComplete(state map[string]*client.JobEvent) bool {
	for _, event := range state {
		processing := false
		if event == nil ||
			event.Job.Status == ffs.InProgress ||
			event.Job.Status == ffs.Queued {
			processing = true
		}
		if processing && event != nil && event.Err == nil {
			return false
		}
	}
	return true
}

func displayName(s ffs.JobStatus) string {
	name, ok := ffs.JobStatusStr[s]
	if !ok {
		return "Unknown"
	}
	return name
}
