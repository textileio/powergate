package client

import (
	"context"
	"time"

	"github.com/textileio/fil-tools/index/miner"
	pb "github.com/textileio/fil-tools/index/miner/pb"
)

// Miners provides an API for viewing miner data
type Miners struct {
	client pb.APIClient
}

// Get returns the current index of available asks
func (a *Miners) Get(ctx context.Context) (*miner.Index, error) {
	reply, err := a.client.Get(ctx, &pb.GetRequest{})
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

	power := make(map[string]miner.Power, len(reply.GetIndex().GetChain().GetPower()))
	for key, val := range reply.GetIndex().GetChain().GetPower() {
		power[key] = miner.Power{
			Power:    val.GetPower(),
			Relative: float64(val.GetRelative()),
		}
	}

	chainIndex := miner.ChainIndex{
		LastUpdated: reply.GetIndex().GetChain().GetLastUpdated(),
		Power:       power,
	}

	index := &miner.Index{
		Meta:  metaIndex,
		Chain: chainIndex,
	}

	return index, nil
}
