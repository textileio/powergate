package filchain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/lotus/chain/types"
)

// FIlChain is an abstraction of the Filecoin network.
type FilChain struct {
	api API
}

// API interacts with a Lotus full-node
type API interface {
	ChainHead(ctx context.Context) (*types.TipSet, error)
}

// New creates a new deal module
func New(api API) *FilChain {
	return &FilChain{
		api: api,
	}
}

// GetHeight returns the current height of the chain for the targeted Lotus node.
func (lc *FilChain) GetHeight(ctx context.Context) (uint64, error) {
	h, err := lc.api.ChainHead(ctx)
	if err != nil {
		return 0, fmt.Errorf("get head from lotus node: %s", err)
	}
	return uint64(h.Height()), nil
}
