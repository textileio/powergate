package stub

import (
	"context"
	"time"

	"github.com/textileio/fil-tools/index/miner/types"
)

// Miners provides an API for viewing miner data
type Miners struct {
}

// Get returns the current index of available asks
func (a *Miners) Get(ctx context.Context) (*types.Index, error) {
	time.Sleep(time.Second * 3)
	info := map[string]types.Meta{
		"miner1": types.Meta{
			LastUpdated: time.Now(),
			UserAgent:   "miner1agent",
			Location: types.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner2": types.Meta{
			LastUpdated: time.Now(),
			UserAgent:   "miner2agent",
			Location: types.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner3": types.Meta{
			LastUpdated: time.Now(),
			UserAgent:   "miner3agent",
			Location: types.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
		"miner4": types.Meta{
			LastUpdated: time.Now(),
			UserAgent:   "miner4agent",
			Location: types.Location{
				Country:   "US",
				Longitude: 111.1,
				Latitude:  45.34,
			},
			Online: true,
		},
	}

	metaIndex := types.MetaIndex{
		Online:  4,
		Offline: 11,
		Info:    info,
	}

	power := map[string]types.Power{
		"miner1": types.Power{
			Power:    88,
			Relative: 0.5,
		},
		"miner2": types.Power{
			Power:    46,
			Relative: 0.34,
		},
		"miner3": types.Power{
			Power:    3,
			Relative: 0.84,
		},
		"miner4": types.Power{
			Power:    234,
			Relative: 0.14,
		},
	}

	chainIndex := types.ChainIndex{
		LastUpdated: 2134567,
		Power:       power,
	}

	index := &types.Index{
		Meta:  metaIndex,
		Chain: chainIndex,
	}

	return index, nil
}
