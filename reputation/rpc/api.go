package rpc

import (
	"context"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/powergate/reputation"
)

// Service implements the gprc service
type Service struct {
	UnimplementedAPIServer

	module *reputation.Module
}

// NewService is a helper to create a new Service
func NewService(m *reputation.Module) *Service {
	return &Service{
		module: m,
	}
}

// AddSource calls Module.AddSource
func (s *Service) AddSource(ctx context.Context, req *AddSourceRequest) (*AddSourceReply, error) {
	maddr, err := ma.NewMultiaddr(req.GetMaddr())
	if err != nil {
		return nil, err
	}
	if err = s.module.AddSource(req.GetId(), maddr); err != nil {
		return nil, err
	}
	return &AddSourceReply{}, nil
}

// GetTopMiners calls Module.GetTopMiners
func (s *Service) GetTopMiners(ctx context.Context, req *GetTopMinersRequest) (*GetTopMinersReply, error) {
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
	return &GetTopMinersReply{TopMiners: pbMinerScores}, nil
}
