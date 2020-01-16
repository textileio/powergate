package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/gosuri/uilive"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/textileio/filecoin/api/client"
	"github.com/textileio/filecoin/deals"
)

// var DealStates = []string{
// 	"DealUnknown",
// 	"DealRejected",
// 	"DealAccepted",
// 	"DealStaged",
// 	"DealSealing",
// 	"DealFailed",
// 	"DealComplete",
// 	"DealError",
// }

var completeStates = []string{
	"DealRejected",
	"DealFailed",
	"DealComplete",
	"DealError",
}

func init() {
	watchCmd.Flags().StringSliceP("cids", "c", []string{}, "List of deal cids to watch")

	rootCmd.AddCommand(watchCmd)
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

		state := make(map[string]*client.WatchEvent, len(cids))
		for _, cid := range cids {
			state[cid.String()] = nil
		}

		writer := uilive.New()
		writer.Start()

		updateOutput(writer, state)

		// ch, err := fcClient.Deals.Watch(ctx, cids)
		// checkErr(err)
		cmd.Println(ctx)
		ch := make(chan client.WatchEvent)
		go driveChannel(ch)

		for {
			event, ok := <-ch
			if ok == false {
				break
			}
			state[event.Deal.ProposalCid.String()] = &event
			updateOutput(writer, state)
			if isComplete(state) {
				break
			}
		}

		writer.Stop()
	},
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
	RenderTable(writer, []string{"miner", "state"}, data)
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

func driveChannel(ch chan client.WatchEvent) {
	cid1, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPU")
	cid2, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPN")

	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid1,
			StateName:   "DealUnknown",
		},
	}
	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid2,
			StateName:   "DealUnknown",
		},
	}

	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid1,
			StateName:   "DealAccepted",
		},
	}
	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid2,
			StateName:   "DealAccepted",
		},
	}

	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid1,
			StateName:   "DealStaged",
		},
	}
	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid2,
			StateName:   "DealStaged",
		},
	}

	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid1,
			StateName:   "DealSealing",
		},
	}
	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid2,
			StateName:   "DealSealing",
		},
	}

	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid1,
			StateName:   "DealComplete",
		},
	}
	time.Sleep(time.Second * 1)
	ch <- client.WatchEvent{
		Deal: deals.DealInfo{
			ProposalCid: cid2,
			StateName:   "DealComplete",
		},
	}
}
