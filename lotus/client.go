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
	"go.opencensus.io/stats/view"
)

var (
	heightProbingInterval = time.Second * 10
	log                   = logging.Logger("lotus-client")
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
		var err error
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

// MonitorLotusSync fires a goroutine that will generate
// metrics with Lotus node height.
func MonitorLotusSync(clientBuilder ClientBuilder) {
	if err := view.Register(vHeight); err != nil {
		log.Fatalf("register metrics views: %v", err)
	}
	go func() {
		for {
			refreshHeightMetric(clientBuilder)
			time.Sleep(heightProbingInterval)
		}
	}()
}

func refreshHeightMetric(clientBuilder ClientBuilder) {
	c, cls, err := clientBuilder()
	if err != nil {
		log.Error("creating lotus client for monitoring: %s", err)
		return
	}
	defer cls()
	heaviest, err := c.ChainHead(context.Background())
	if err != nil {
		log.Errorf("get lotus sync status: %s", err)
		return
	}
	stats.Record(context.Background(), mLotusHeight.M(int64(heaviest.Height())))
}
