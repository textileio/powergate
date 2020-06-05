package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/miner"
)

// Miners provides an API for viewing miner data.
type Miners struct {
}

// Get returns the current index of available asks.
func (a *Miners) Get(ctx context.Context) (*miner.IndexSnapshot, error) {
	time.Sleep(time.Second * 3)
	info := map[string]miner.Meta{
		"miner1": {
			LastUpdated: time.Now(),
			UserAgent:   "miner1agent",
			Location: miner.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner2": {
			LastUpdated: time.Now(),
			UserAgent:   "miner2agent",
			Location: miner.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner3": {
			LastUpdated: time.Now(),
			UserAgent:   "miner3agent",
			Location: miner.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner4": {
			LastUpdated: time.Now(),
			UserAgent:   "miner4agent",
			Location: miner.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
	}

	metaIndex := miner.MetaIndex{
		Online:  4,
		Offline: 11,
		Info:    info,
	}

	miners := map[string]miner.OnChainData{
		"miner1": {
			Power:         88,
			RelativePower: 0.5,
			SectorSize:    512,
			ActiveDeals:   10,
		},
		"miner2": {
			Power:         46,
			RelativePower: 0.34,
			SectorSize:    512,
			ActiveDeals:   12,
		},
		"miner3": {
			Power:         3,
			RelativePower: 0.84,
			SectorSize:    1024,
			ActiveDeals:   1,
		},
		"miner4": {
			Power:         234,
			RelativePower: 0.14,
			SectorSize:    512,
			ActiveDeals:   3,
		},
	}

	chainIndex := miner.ChainIndex{
		LastUpdated: 2134567,
		Miners:      miners,
	}

	index := &miner.IndexSnapshot{
		Meta:    metaIndex,
		OnChain: chainIndex,
	}

	return index, nil
}
