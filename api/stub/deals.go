package stub

import (
	"context"
	"io"
	"time"

	cid "github.com/ipfs/go-cid"
	"github.com/textileio/filecoin/api/client"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/index/ask"
)

// Deals provides an API for managing deals and storing data
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

// AvailableAsks executes a query to retrieve active Asks
func (d *Deals) AvailableAsks(ctx context.Context, query ask.Query) ([]ask.StorageAsk, error) {
	time.Sleep(time.Second * 3)
	var asks = []ask.StorageAsk{
		ask.StorageAsk{
			Miner:        "asdfzxvc123",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    1123456,
			Expiry:       12345677,
		},
		ask.StorageAsk{
			Miner:        "xcvbfgfhh657",
			Price:        3420,
			MinPieceSize: 2048,
			Timestamp:    89798,
			Expiry:       12345677,
		},
		ask.StorageAsk{
			Miner:        "asdfzxvc123",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    1123456,
			Expiry:       12345677,
		},
		ask.StorageAsk{
			Miner:        "asdfzxvc123",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    1123456,
			Expiry:       12345677,
		},
	}
	return asks, nil
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs
func (d *Deals) Store(ctx context.Context, addr string, data io.Reader, dealConfigs []deals.DealConfig, duration uint64) ([]cid.Cid, []deals.DealConfig, error) {
	time.Sleep(time.Second * 3)
	success := []cid.Cid{}
	failed := []deals.DealConfig{}
	for i, dealConfig := range dealConfigs {
		if i == 0 {
			failed = append(failed, dealConfig)
		} else {
			success = append(success, cids[i-1])
		}
	}
	return success, failed, nil
}

// Watch returnas a channel with state changes of indicated proposals
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
				Deal: deals.DealInfo{
					ProposalCid: proposal,
					StateName:   dealState,
				},
			}
		}
	}
	close(ch)
}
