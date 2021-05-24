package lotus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/textileio/powergate/v2/util"
)

var (
	log = logging.Logger("lotus-client")
)

// ClientBuilder creates a new Lotus client.
type ClientBuilder func(ctx context.Context) (*api.FullNodeStruct, func(), error)

// NewBuilder creates a new ClientBuilder.
func NewBuilder(maddr ma.Multiaddr, authToken string, connRetries int) (ClientBuilder, error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, err
	}
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}

	return func(ctx context.Context) (*api.FullNodeStruct, func(), error) {
		var api api.FullNodeStruct
		var closer jsonrpc.ClientCloser
		var err error
		for i := 0; i < connRetries; i++ {
			if ctx.Err() != nil {
				return nil, nil, fmt.Errorf("canceled by context")
			}
			closer, err = jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin",
				[]interface{}{
					&api.Internal,
					&api.CommonStruct.Internal,
				}, headers)
			if err == nil {
				break
			}
			log.Warnf("failed to connect to Lotus client %s, retrying...", err)
			time.Sleep(time.Second * 10)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't connect to Lotus API: %s", err)
		}

		return &api, closer, nil
	}, nil
}
