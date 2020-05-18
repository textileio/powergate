package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/ask"
	"github.com/textileio/powergate/index/ask/runner"
)

// Asks stub
type Asks struct {
}

// Get returns the current index of available asks
func (a *Asks) Get(ctx context.Context) (*ask.Index, error) {
	time.Sleep(time.Second * 3)
	storage := map[string]ask.StorageAsk{
		"miner1": {
			Miner:        "miner1",
			Price:        5001,
			MinPieceSize: 1004,
			Timestamp:    1,
			Expiry:       100,
		},
		"miner2": {
			Miner:        "miner2",
			Price:        5002,
			MinPieceSize: 1004,
			Timestamp:    2,
			Expiry:       100,
		},
		"miner3": {
			Miner:        "miner3",
			Price:        5003,
			MinPieceSize: 1004,
			Timestamp:    3,
			Expiry:       100,
		},
		"miner4": {
			Miner:        "miner4",
			Price:        5004,
			MinPieceSize: 1004,
			Timestamp:    4,
			Expiry:       100,
		},
	}
	return &ask.Index{
		LastUpdated:        time.Now(),
		StorageMedianPrice: 5000,
		Storage:            storage,
	}, nil
}

// Query executes a query to retrieve active Asks
func (a *Asks) Query(ctx context.Context, query runner.Query) ([]ask.StorageAsk, error) {
	time.Sleep(time.Second * 3)
	var asks = []ask.StorageAsk{
		ask.StorageAsk{
			Miner:        "miner1",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    1,
			Expiry:       100,
		},
		ask.StorageAsk{
			Miner:        "miner2",
			Price:        3420,
			MinPieceSize: 2048,
			Timestamp:    2,
			Expiry:       200,
		},
		ask.StorageAsk{
			Miner:        "miner3",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    3,
			Expiry:       100,
		},
		ask.StorageAsk{
			Miner:        "miner4",
			Price:        1245,
			MinPieceSize: 1024,
			Timestamp:    4,
			Expiry:       200,
		},
	}
	return asks, nil
}
