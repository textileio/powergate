package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/apoorvam/goterminal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
	proto "github.com/textileio/powergate/proto/powergate/v1"
)

func init() {
	storageJobsCmd.AddCommand(storageJobsWatchCmd)
}

var storageJobsWatchCmd = &cobra.Command{
	Use:   "watch [jobid,...]",
	Short: "Watch for storage job status updates",
	Long:  `Watch for storage job status updates`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jobIds := strings.Split(args[0], ",")
		watchJobIds(jobIds...)
	},
}

func watchJobIds(jobIds ...string) {
	state := make(map[string]*client.WatchStorageJobsEvent, len(jobIds))
	for _, jobID := range jobIds {
		state[jobID] = nil
	}

	writer := goterminal.New(os.Stdout)

	updateJobsOutput(writer, state)

	ch := make(chan client.WatchStorageJobsEvent)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := powClient.StorageJobs.Watch(mustAuthCtx(ctx), ch, jobIds...)
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
		state[event.Res.StorageJob.Id] = &event
		updateJobsOutput(writer, state)
		if jobsComplete(state) {
			break
		}
	}
}

func updateJobsOutput(writer *goterminal.Writer, state map[string]*client.WatchStorageJobsEvent) {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var data [][]string
	for _, k := range keys {
		if state[k] != nil {
			var val string
			if state[k].Res.StorageJob.Status == proto.JobStatus_JOB_STATUS_FAILED {
				val = fmt.Sprintf("%v %v", state[k].Res.StorageJob.Status.String(), state[k].Res.StorageJob.ErrorCause)
			} else if state[k].Err != nil {
				val = fmt.Sprintf("Error: %v", state[k].Err.Error())
			} else {
				val = state[k].Res.StorageJob.Status.String()
			}
			data = append(data, []string{k, val, "", "", ""})
			for _, dealInfo := range state[k].Res.StorageJob.DealInfo {
				data = append(data, []string{"", "", dealInfo.Miner, strconv.FormatUint(dealInfo.PricePerEpoch, 10), dealInfo.StateName})
			}
		} else {
			data = append(data, []string{k, "awaiting state", "", "", ""})
		}
	}

	RenderTable(writer, []string{"Job id", "Status", "Miner", "Price", "Deal Status"}, data)

	writer.Clear()
	_ = writer.Print()
}

func jobsComplete(state map[string]*client.WatchStorageJobsEvent) bool {
	for _, event := range state {
		processing := false
		if event == nil ||
			event.Res.StorageJob.Status == proto.JobStatus_JOB_STATUS_EXECUTING ||
			event.Res.StorageJob.Status == proto.JobStatus_JOB_STATUS_QUEUED {
			processing = true
		}
		if processing && event != nil && event.Err == nil {
			return false
		}
	}
	return true
}
