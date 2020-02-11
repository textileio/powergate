package stub

import (
	"context"
	"time"

	"github.com/textileio/fil-tools/index/ask/types"
)

// Asks stub
type Asks struct {
}

// Get returns the current index of available asks
func (a *Asks) Get(ctx context.Context) (*types.Index, error) {
	time.Sleep(time.Second * 3)
	storage := map[string]types.StorageAsk{
		"miner1": types.StorageAsk{
			Miner:        "miner1",
			Price:        5001,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner2": types.StorageAsk{
			Miner:        "miner2",
			Price:        5002,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner3": types.StorageAsk{
			Miner:        "miner3",
			Price:        5003,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner4": types.StorageAsk{
			Miner:        "miner4",
			Price:        5004,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
	}
	return &types.Index{
		LastUpdated:        time.Now(),
		StorageMedianPrice: 5000,
		Storage:            storage,
	}, nil
}

// Query executes a query to retrieve active Asks
func (a *Asks) Query(ctx context.Context, query types.Query) ([]types.StorageAsk, error) {
	time.Sleep(time.Second * 3)
	var asks = []types.StorageAsk{
		types.StorageAsk{
			Miner:        "miner1",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		types.StorageAsk{
			Miner:        "miner2",
			Price:        3420,
			MinPieceSize: 2048,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		types.StorageAsk{
			Miner:        "miner3",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		types.StorageAsk{
			Miner:        "miner4",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
	}
	return asks, nil
}
