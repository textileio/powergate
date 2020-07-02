package rpc

import (
	"context"

	"github.com/textileio/powergate/index/miner"
)

// RPC implements the gprc service.
type RPC struct {
	UnimplementedRPCServiceServer

	module miner.Module
}

// New creates a new rpc service.
func New(mi miner.Module) *RPC {
	return &RPC{
		module: mi,
	}
}

// Get calls miner index Get.
func (s *RPC) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	index := s.module.Get()

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

	pbPower := make(map[string]*OnChainData, len(index.OnChain.Miners))
	for key, miner := range index.OnChain.Miners {
		pbPower[key] = &OnChainData{
			Power:         miner.Power,
			RelativePower: float32(miner.RelativePower),
			SectorSize:    miner.SectorSize,
			ActiveDeals:   miner.ActiveDeals,
		}
	}

	meta := &MetaIndex{
		Online:  index.Meta.Online,
		Offline: index.Meta.Offline,
		Info:    info,
	}

	chain := &OnChainIndex{
		LastUpdated: index.OnChain.LastUpdated,
		Miners:      pbPower,
	}

	pbIndex := &Index{
		Meta:  meta,
		Chain: chain,
	}

	return &GetResponse{Index: pbIndex}, nil
}
