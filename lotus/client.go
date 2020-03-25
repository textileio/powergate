package lotus

import (
	"context"
	"net/http"
	"time"

	"github.com/filecoin-project/lotus/lib/jsonrpc"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/lotus-client/api/apistruct"

	"github.com/textileio/powergate/util"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	lotusSyncStatusInterval = time.Second * 10
	log                     = logging.Logger("deals")
)

// New creates a new client to Lotus API
func New(maddr ma.Multiaddr, authToken string) (*apistruct.FullNodeStruct, func(), error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, nil, err
	}
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}
	var api apistruct.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient("ws://"+addr+"/rpc/v0", "Filecoin",
		[]interface{}{
			&api.Internal,
			&api.CommonStruct.Internal,
		}, headers)
	if err != nil {
		return nil, nil, err
	}

	if err := view.Register(vHeight); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go monitorLotusSync(ctx, &api)

	return &api, func() {
		cancel()
		closer()
	}, nil

}

func monitorLotusSync(ctx context.Context, c *apistruct.FullNodeStruct) {
	refreshHeightMetric(ctx, c)
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing lotus sync monitor")
			return
		case <-time.After(lotusSyncStatusInterval):
			refreshHeightMetric(ctx, c)
		}
	}
}

func refreshHeightMetric(ctx context.Context, c *apistruct.FullNodeStruct) {
	heaviest, err := c.ChainHead(ctx)
	if err != nil {
		log.Errorf("error when getting lotus sync status: %s", err)
		return
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height())))
}
