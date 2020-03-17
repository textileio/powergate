package lotuschain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/powergate/ffs/filcold"
)

type LotusChain struct {
	api API
}

var _ filcold.FilChain = (*LotusChain)(nil)

// API interacts with a Lotus full-node
type API interface {
	ChainHead(ctx context.Context) (*types.TipSet, error)
}

// New creates a new deal module
func New(api API) *LotusChain {
	return &LotusChain{
		api: api,
	}
}

func (lc *LotusChain) GetHeight(ctx context.Context) (uint64, error) {
	h, err := lc.api.ChainHead(ctx)
	if err != nil {
		return 0, fmt.Errorf("get head from lotus node: %s", err)
	}
	return h.Height(), nil
}
