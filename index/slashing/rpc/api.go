package rpc

import (
	"context"

	"github.com/textileio/powergate/index/slashing"
)

// Service implements the gprc service
type Service struct {
	UnimplementedAPIServer

	index *slashing.Index
}

// NewService is a helper to create a new Service
func NewService(si *slashing.Index) *Service {
	return &Service{
		index: si,
	}
}

// Get calls slashing index Get
func (s *Service) Get(ctx context.Context, req *GetRequest) (*GetReply, error) {
	index := s.index.Get()

	miners := make(map[string]*Slashes, len(index.Miners))
	for key, slashes := range index.Miners {
		miners[key] = &Slashes{
			Epochs: slashes.Epochs,
		}
	}

	pbIndex := &Index{
		TipSetKey: index.TipSetKey,
		Miners:    miners,
	}

	return &GetReply{Index: pbIndex}, nil
}
