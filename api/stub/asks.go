package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/ask"
)

// Asks stub
type Asks struct {
}

// Get returns the current index of available asks
func (a *Asks) Get(ctx context.Context) (*ask.Index, error) {
	time.Sleep(time.Second * 3)
	storage := map[string]ask.StorageAsk{
		"miner1": ask.StorageAsk{
			Miner:        "miner1",
			Price:        5001,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner2": ask.StorageAsk{
			Miner:        "miner2",
			Price:        5002,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner3": ask.StorageAsk{
			Miner:        "miner3",
			Price:        5003,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		"miner4": ask.StorageAsk{
			Miner:        "miner4",
			Price:        5004,
			MinPieceSize: 1004,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
	}
	return &ask.Index{
		LastUpdated:        time.Now(),
		StorageMedianPrice: 5000,
		Storage:            storage,
	}, nil
}

// Query executes a query to retrieve active Asks
func (a *Asks) Query(ctx context.Context, query ask.Query) ([]ask.StorageAsk, error) {
	time.Sleep(time.Second * 3)
	var asks = []ask.StorageAsk{
		ask.StorageAsk{
			Miner:        "miner1",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		ask.StorageAsk{
			Miner:        "miner2",
			Price:        3420,
			MinPieceSize: 2048,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		ask.StorageAsk{
			Miner:        "miner3",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
		ask.StorageAsk{
			Miner:        "miner4",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    uint64(time.Now().Unix()),
			Expiry:       uint64(time.Now().Unix()),
		},
	}
	return asks, nil
}
