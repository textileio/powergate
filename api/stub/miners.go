package stub

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/miner"
)

// Miners provides an API for viewing miner data
type Miners struct {
}

// Get returns the current index of available asks
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

	power := map[string]miner.Power{
		"miner1": miner.Power{
			Power:    88,
			Relative: 0.5,
		},
		"miner2": miner.Power{
			Power:    46,
			Relative: 0.34,
		},
		"miner3": miner.Power{
			Power:    3,
			Relative: 0.84,
		},
		"miner4": miner.Power{
			Power:    234,
			Relative: 0.14,
		},
	}

	chainIndex := miner.ChainIndex{
		LastUpdated: 2134567,
		Power:       power,
	}

	index := &miner.IndexSnapshot{
		Meta:  metaIndex,
		Chain: chainIndex,
	}

	return index, nil
}
