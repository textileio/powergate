package rpc

import (
	"context"

	"github.com/textileio/powergate/index/miner"
)

// RPC implements the gprc service
type RPC struct {
	UnimplementedRPCServiceServer

	index *miner.Index
}

// New creates a new rpc service
func New(mi *miner.Index) *RPC {
	return &RPC{
		index: mi,
	}
}

// Get calls miner index Get
func (s *RPC) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	index := s.index.Get()

	info := make(map[string]*Meta, len(index.Meta.Info))
	for key, meta := range index.Meta.Info {
		location := &Location{
			Country:   meta.Location.Country,
			Longitude: meta.Location.Longitude,
			Latitude:  meta.Location.Latitude,
		}
		info[key] = &Meta{
			LastUpdated: meta.LastUpdated.Unix(),
			UserAgent:   meta.UserAgent,
			Location:    location,
			Online:      meta.Online,
		}
	}

	pbPower := make(map[string]*Power, len(index.Chain.Power))
	for key, power := range index.Chain.Power {
		pbPower[key] = &Power{
			Power:    power.Power,
			Relative: float32(power.Relative),
		}
	}

	meta := &MetaIndex{
		Online:  index.Meta.Online,
		Offline: index.Meta.Offline,
		Info:    info,
	}

	chain := &ChainIndex{
		LastUpdated: index.Chain.LastUpdated,
		Power:       pbPower,
	}

	pbIndex := &Index{
		Meta:  meta,
		Chain: chain,
	}

	return &GetResponse{Index: pbIndex}, nil
}
