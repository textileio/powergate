package miner

import (
	"context"

	pb "github.com/textileio/fil-tools/index/miner/pb"
)

// Service implements the gprc service
type Service struct {
	pb.UnimplementedAPIServer

	index *MinerIndex
}

// NewService is a helper to create a new Service
func NewService(mi *MinerIndex) *Service {
	return &Service{
		index: mi,
	}
}

// Get calls miner index Get
func (s *Service) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetReply, error) {
	index := s.index.Get()

	info := make(map[string]*pb.Meta, len(index.Meta.Info))
	for key, meta := range index.Meta.Info {
		location := &pb.Location{
			Country:   meta.Location.Country,
			Longitude: meta.Location.Longitude,
			Latitude:  meta.Location.Latitude,
		}
		info[key] = &pb.Meta{
			LastUpdated: meta.LastUpdated.Unix(),
			UserAgent:   meta.UserAgent,
			Location:    location,
			Online:      meta.Online,
		}
	}

	pbPower := make(map[string]*pb.Power, len(index.Chain.Power))
	for key, power := range index.Chain.Power {
		pbPower[key] = &pb.Power{
			Power:    power.Power,
			Relative: float32(power.Relative),
		}
	}

	meta := &pb.MetaIndex{
		Online:  index.Meta.Online,
		Offline: index.Meta.Offline,
		Info:    info,
	}

	chain := &pb.ChainIndex{
		LastUpdated: index.Chain.LastUpdated,
		Power:       pbPower,
	}

	pbIndex := &pb.Index{
		Meta:  meta,
		Chain: chain,
	}

	return &pb.GetReply{Index: pbIndex}, nil
}
