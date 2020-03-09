package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/gosuri/uilive"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/textileio/powergate/api/client"
)

func init() {
	watchCmd.Flags().StringSliceP("cids", "c", []string{}, "List of deal cids to watch")

	dealsCmd.AddCommand(watchCmd)
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch storage process",
	Long:  `Watch storage process`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		cidStrings, err := cmd.Flags().GetStringSlice("cids")
		checkErr(err)
		if len(cidStrings) == 0 {
			Fatal(errors.New("you must provide at least one cid using the cids flag"))
		}

		cids := make([]cid.Cid, len(cidStrings))
		for i, s := range cidStrings {
			cid, err := cid.Parse(s)
			checkErr(err)
			cids[i] = cid
		}

		processWatchEventsUntilDone(ctx, cids)
	},
}

func processWatchEventsUntilDone(ctx context.Context, cids []cid.Cid) {
	state := make(map[string]*client.WatchEvent, len(cids))
	for _, cid := range cids {
		state[cid.String()] = nil
	}

	writer := uilive.New()
	writer.Start()

	updateOutput(writer, state)

	ch, err := fcClient.Deals.Watch(ctx, cids)
	checkErr(err)

	for {
		event, ok := <-ch
		if !ok {
			break
		}
		state[event.Deal.ProposalCid.String()] = &event
		updateOutput(writer, state)
		if isComplete(state) {
			break
		}
	}

	writer.Stop()
}

func updateOutput(writer io.Writer, state map[string]*client.WatchEvent) {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := make([][]string, len(keys))
	for i, k := range keys {
		val := "awaiting state"
		if state[k] != nil {
			if state[k].Deal.StateName == "DealError" && state[k].Err != nil {
				val = fmt.Sprintf("%v: %v", state[k].Deal.StateName, state[k].Err.Error())
			} else if state[k].Err != nil {
				val = fmt.Sprintf("Error: %v", state[k].Err.Error())
			} else {
				val = state[k].Deal.StateName
			}
		}
		data[i] = []string{
			k,
			val,
		}
	}
	RenderTable(writer, []string{"deal id", "state"}, data)
}

func isComplete(state map[string]*client.WatchEvent) bool {
	for _, value := range state {
		processing := false
		if value == nil ||
			value.Deal.StateName == "DealUnknown" ||
			value.Deal.StateName == "DealAccepted" ||
			value.Deal.StateName == "DealStaged" ||
			value.Deal.StateName == "DealSealing" {
			processing = true
		}
		if processing && value != nil && value.Err == nil {
			return false
		}
	}
	return true
}
