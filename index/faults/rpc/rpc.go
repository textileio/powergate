package rpc

import (
	"context"

	"github.com/textileio/powergate/index/faults"
)

// RPC implements the gprc service.
type RPC struct {
	UnimplementedRPCServiceServer

	module faults.Module
}

// New creates a new rpc service.
func New(fi faults.Module) *RPC {
	return &RPC{
		module: fi,
	}
}

// Get calls faults index Get.
func (s *RPC) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	index := s.module.Get()

	miners := make(map[string]*Faults, len(index.Miners))
	for key, faults := range index.Miners {
		miners[key] = &Faults{
			Epochs: faults.Epochs,
		}
	}

	pbIndex := &Index{
		Tipsetkey: index.TipSetKey,
		Miners:    miners,
	}

	return &GetResponse{Index: pbIndex}, nil
}
