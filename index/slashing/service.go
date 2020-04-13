package slashing

import (
	"context"

	pb "github.com/textileio/powergate/index/slashing/pb"
)

// Service implements the gprc service
type Service struct {
	pb.UnimplementedAPIServer

	index *Index
}

// NewService is a helper to create a new Service
func NewService(si *Index) *Service {
	return &Service{
		index: si,
	}
}

// Get calls slashing index Get
func (s *Service) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetReply, error) {
	index := s.index.Get()

	miners := make(map[string]*pb.Slashes, len(index.Miners))
	for key, slashes := range index.Miners {
		miners[key] = &pb.Slashes{
			Epochs: slashes.Epochs,
		}
	}

	pbIndex := &pb.Index{
		TipSetKey: index.TipSetKey,
		Miners:    miners,
	}

	return &pb.GetReply{Index: pbIndex}, nil
}
