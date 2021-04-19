package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/apoorvam/goterminal"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/v2/api/client"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
)

var (
	// PowClient is the powergate client.
	PowClient *client.Client

	// CmdTimeout is the standard timeout.
	CmdTimeout = time.Second * 60
)

// FmtOutput allows to configure where Message(), Success(), and
// Fatal() helpers should write.
var FmtOutput = os.Stdout

// Message prints a message to stdout.
func Message(format string, args ...interface{}) {
	fmt.Fprintln(FmtOutput, aurora.Sprintf(aurora.BrightBlack("> "+format), args...))
}

// Success prints a success message to stdout.
func Success(format string, args ...interface{}) {
	fmt.Fprintln(FmtOutput, aurora.Sprintf(aurora.Cyan("> Success! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

// Fatal prints a fatal error to stdout, and exits immediately with
// error code 1.
func Fatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Fprintln(FmtOutput, aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
	os.Exit(1)
}

// RenderTable renders a table with header columns and data rows to writer.
func RenderTable(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorder(false)
	headersColors := make([]tablewriter.Colors, len(header))
	for i := range headersColors {
		headersColors[i] = tablewriter.Colors{tablewriter.FgHiBlackColor}
	}
	table.SetHeaderColor(headersColors...)
	table.AppendBulk(data)
	table.Render()
}

// CheckErr calls Fatal if there is an error.
func CheckErr(e error) {
	if e != nil {
		Fatal(e)
	}
}

// WatchJobIds watches jobs.
func WatchJobIds(jobIds ...string) {
	state := make(map[string]*client.WatchStorageJobsEvent, len(jobIds))
	for _, jobID := range jobIds {
		state[jobID] = nil
	}

	writer := goterminal.New(os.Stdout)

	updateJobsOutput(writer, state)

	ch := make(chan client.WatchStorageJobsEvent)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := PowClient.StorageJobs.Watch(MustAuthCtx(ctx), ch, jobIds...)
	CheckErr(err)

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
			if state[k].Res.StorageJob.Status == userPb.JobStatus_JOB_STATUS_FAILED {
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
			event.Res.StorageJob.Status == userPb.JobStatus_JOB_STATUS_EXECUTING ||
			event.Res.StorageJob.Status == userPb.JobStatus_JOB_STATUS_QUEUED {
			processing = true
		}
		if processing && event != nil && event.Err == nil {
			return false
		}
	}
	return true
}

// MustAuthCtx returns the auth context built from viper token.
func MustAuthCtx(ctx context.Context) context.Context {
	token := viper.GetString("token")
	if token == "" {
		Fatal(errors.New("must provide -t token"))
	}
	return context.WithValue(ctx, client.AuthKey, token)
}

// AdminAuthCtx returns the admin auth context built from viper admin token.
func AdminAuthCtx(ctx context.Context) context.Context {
	token := viper.GetString("admin-token")
	return context.WithValue(ctx, client.AdminKey, token)
}
