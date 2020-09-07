package lotus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api/apistruct"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/textileio/powergate/util"
	"go.opencensus.io/stats"
)

var (
	lotusSyncStatusInterval = time.Second * 10
	log                     = logging.Logger("lotus-client")
)

// ClientBuilder creates a new Lotus client.
type ClientBuilder func() (*apistruct.FullNodeStruct, func(), error)

// NewBuilder creates a new ClientBuilder.
func NewBuilder(maddr ma.Multiaddr, authToken string, connRetries int) (ClientBuilder, error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, err
	}
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}

	return func() (*apistruct.FullNodeStruct, func(), error) {
		var api apistruct.FullNodeStruct
		var closer jsonrpc.ClientCloser
		for i := 0; i < connRetries; i++ {
			closer, err = jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin",
				[]interface{}{
					&api.Internal,
					&api.CommonStruct.Internal,
				}, headers)
			if err == nil {
				break
			}
			log.Warnf("failed to connect to Lotus client %s, retrying...", err)
			time.Sleep(time.Second * 5)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't connect to Lotus API: %s", err)
		}

		return &api, closer, nil
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
