package rpc

import (
	"context"

	"github.com/textileio/powergate/index/slashing"
)

// RPC implements the gprc service
type RPC struct {
	UnimplementedRPCServiceServer

	index *slashing.Index
}

// New creates a new rpc service
func New(si *slashing.Index) *RPC {
	return &RPC{
		index: si,
	}
}

// Get calls slashing index Get
func (s *RPC) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	index := s.index.Get()

	miners := make(map[string]*Slashes, len(index.Miners))
	for key, slashes := range index.Miners {
		miners[key] = &Slashes{
			Epochs: slashes.Epochs,
		}
	}

	pbIndex := &Index{
		Tipsetkey: index.TipSetKey,
		Miners:    miners,
	}

	return &GetResponse{Index: pbIndex}, nil
}
