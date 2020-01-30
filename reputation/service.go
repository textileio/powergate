package reputation

import (
	"context"

	ma "github.com/multiformats/go-multiaddr"
	pb "github.com/textileio/filecoin/reputation/pb"
)

// Service implements the gprc service
type Service struct {
	pb.UnimplementedAPIServer

	module *Module
}

// NewService is a helper to create a new Service
func NewService(m *Module) *Service {
	return &Service{
		module: m,
	}
}

// AddSource calls Module.AddSource
func (s *Service) AddSource(ctx context.Context, req *pb.AddSourceRequest) (*pb.AddSourceReply, error) {
	maddr, err := ma.NewMultiaddr(req.GetMaddr())
	if err != nil {
		return nil, err
	}
	if err = s.module.AddSource(req.GetId(), maddr); err != nil {
		return nil, err
	}
	return &pb.AddSourceReply{}, nil
}

// GetTopMiners calls Module.GetTopMiners
func (s *Service) GetTopMiners(ctx context.Context, req *pb.GetTopMinersRequest) (*pb.GetTopMinersReply, error) {
	minerScores, err := s.module.GetTopMiners(int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	pbMinerScores := make([]*pb.MinerScore, len(minerScores))
	for i, minerScore := range minerScores {
		pbMinerScores[i] = &pb.MinerScore{
			Addr:  minerScore.Addr,
			Score: int32(minerScore.Score),
		}
	}
	return &pb.GetTopMinersReply{TopMiners: pbMinerScores}, nil
}
