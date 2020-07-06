package stub

import (
	"context"
	"io"
	"strings"
	"time"

	cid "github.com/ipfs/go-cid"
	"github.com/textileio/powergate/api/client"
	"github.com/textileio/powergate/deals"
)

// Deals provides an API for managing deals and storing data.
type Deals struct {
}

var cids = []cid.Cid{}
var dealStates = []string{
	"DealUnknown",
	"DealAccepted",
	"DealStaged",
	"DealSealing",
	"DealComplete",
}

func init() {
	cid1, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPA")
	cid2, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPB")
	cid3, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPC")
	cid4, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPD")
	cid5, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPE")
	cid6, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPF")
	cid7, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPG")
	cid8, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPH")
	cid9, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPI")
	cid10, _ := cid.Parse("QmY7Yh4UquoXHLPFo2XbhXkhBvFoPwmQUSa92pxnxjQuPJ")
	cids = append(cids, cid1, cid2, cid3, cid4, cid5, cid6, cid7, cid8, cid9, cid10)
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs.
func (d *Deals) Store(ctx context.Context, addr string, data io.Reader, dealConfigs []deals.StorageDealConfig, duration uint64) ([]cid.Cid, []deals.StorageDealConfig, error) {
	time.Sleep(time.Second * 3)
	success := []cid.Cid{}
	failed := []deals.StorageDealConfig{}
	for i, dealConfig := range dealConfigs {
		if i == 0 {
			failed = append(failed, dealConfig)
		} else {
			success = append(success, cids[i-1])
		}
	}
	return success, failed, nil
}

// Watch returnas a channel with state changes of indicated proposals.
func (d *Deals) Watch(ctx context.Context, proposals []cid.Cid) (<-chan client.WatchEvent, error) {
	ch := make(chan client.WatchEvent)
	go driveChannel(ch, proposals)
	return ch, nil
}

func driveChannel(ch chan client.WatchEvent, proposals []cid.Cid) {
	for _, dealState := range dealStates {
		for _, proposal := range proposals {
			time.Sleep(time.Millisecond * 1000)
			ch <- client.WatchEvent{
				Deal: deals.StorageDealInfo{
					ProposalCid: proposal,
					StateName:   dealState,
				},
			}
		}
	}
	close(ch)
}

// Retrieve is used to fetch data from filecoin.
func (d *Deals) Retrieve(ctx context.Context, waddr string, cid cid.Cid) (io.Reader, error) {
	time.Sleep(time.Second * 3)
	return strings.NewReader("hello there"), nil
}
