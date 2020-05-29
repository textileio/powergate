package rpc

import (
	"context"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/reputation"
)

// RPC implements the gprc service
type RPC struct {
	UnimplementedRPCServiceServer

	module *reputation.Module
}

// New creates a new rpc service
func New(m *reputation.Module) *RPC {
	return &RPC{
		module: m,
	}
}

// AddSource calls Module.AddSource
func (s *RPC) AddSource(ctx context.Context, req *AddSourceRequest) (*AddSourceResponse, error) {
	maddr, err := ma.NewMultiaddr(req.GetMaddr())
	if err != nil {
		return nil, err
	}
	if err = s.module.AddSource(req.GetId(), maddr); err != nil {
		return nil, err
	}
	return &AddSourceResponse{}, nil
}

// GetTopMiners calls Module.GetTopMiners
func (s *RPC) GetTopMiners(ctx context.Context, req *GetTopMinersRequest) (*GetTopMinersResponse, error) {
	minerScores, err := s.module.GetTopMiners(int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	pbMinerScores := make([]*MinerScore, len(minerScores))
	for i, minerScore := range minerScores {
		pbMinerScores[i] = &MinerScore{
			Addr:  minerScore.Addr,
			Score: int32(minerScore.Score),
		}
	}
	return &GetTopMinersResponse{TopMiners: pbMinerScores}, nil
}
