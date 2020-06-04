package client

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/miner"
	"github.com/textileio/powergate/index/miner/rpc"
)

// Miners provides an API for viewing miner data
type Miners struct {
	client rpc.RPCServiceClient
}

// Get returns the current index of available asks
func (a *Miners) Get(ctx context.Context) (*miner.IndexSnapshot, error) {
	reply, err := a.client.Get(ctx, &rpc.GetRequest{})
	if err != nil {
		return nil, err
	}

	info := make(map[string]miner.Meta, len(reply.GetIndex().GetMeta().GetInfo()))
	for key, val := range reply.GetIndex().GetMeta().GetInfo() {
		info[key] = miner.Meta{
			LastUpdated: time.Unix(val.GetLastUpdated(), 0),
			UserAgent:   val.GetUserAgent(),
			Location: miner.Location{
				Country:   val.GetLocation().GetCountry(),
				Longitude: val.GetLocation().GetLongitude(),
				Latitude:  val.GetLocation().GetLatitude(),
			},
			Online: val.GetOnline(),
		}
	}

	metaIndex := miner.MetaIndex{
		Online:  reply.GetIndex().GetMeta().GetOnline(),
		Offline: reply.GetIndex().GetMeta().GetOffline(),
		Info:    info,
	}

	miners := make(map[string]miner.OnChainData, len(reply.GetIndex().GetChain().GetMiners()))
	for key, val := range reply.GetIndex().GetChain().GetMiners() {
		miners[key] = miner.OnChainData{
			Power:         val.GetPower(),
			RelativePower: float64(val.GetRelativePower()),
			SectorSize:    uint64(val.GetSectorSize()),
			ActiveDeals:   uint64(val.GetActiveDeals()),
		}
	}

	chainIndex := miner.ChainIndex{
		LastUpdated: reply.GetIndex().GetChain().GetLastUpdated(),
		Miners:      miners,
	}

	index := &miner.IndexSnapshot{
		Meta:    metaIndex,
		OnChain: chainIndex,
	}

	return index, nil
}
