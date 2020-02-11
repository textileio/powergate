package client

import (
	"context"
	"time"

	pb "github.com/textileio/fil-tools/index/miner/pb"
	"github.com/textileio/fil-tools/index/miner/types"
)

// Miners provides an API for viewing miner data
type Miners struct {
	client pb.APIClient
}

// Get returns the current index of available asks
func (a *Miners) Get(ctx context.Context) (*types.Index, error) {
	reply, err := a.client.Get(ctx, &pb.GetRequest{})
	if err != nil {
		return nil, err
	}

	info := make(map[string]types.Meta, len(reply.GetIndex().GetMeta().GetInfo()))
	for key, val := range reply.GetIndex().GetMeta().GetInfo() {
		info[key] = types.Meta{
			LastUpdated: time.Unix(val.GetLastUpdated(), 0),
			UserAgent:   val.GetUserAgent(),
			Location: types.Location{
				Country:   val.GetLocation().GetCountry(),
				Longitude: val.GetLocation().GetLongitude(),
				Latitude:  val.GetLocation().GetLatitude(),
			},
			Online: val.GetOnline(),
		}
	}

	metaIndex := types.MetaIndex{
		Online:  reply.GetIndex().GetMeta().GetOnline(),
		Offline: reply.GetIndex().GetMeta().GetOffline(),
		Info:    info,
	}

	power := make(map[string]types.Power, len(reply.GetIndex().GetChain().GetPower()))
	for key, val := range reply.GetIndex().GetChain().GetPower() {
		power[key] = types.Power{
			Power:    val.GetPower(),
			Relative: float64(val.GetRelative()),
		}
	}

	chainIndex := types.ChainIndex{
		LastUpdated: reply.GetIndex().GetChain().GetLastUpdated(),
		Power:       power,
	}

	index := &types.Index{
		Meta:  metaIndex,
		Chain: chainIndex,
	}

	return index, nil
}
