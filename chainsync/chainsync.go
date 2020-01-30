package chainsync

import (
	"context"

	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

// API provides an abstraction to a Filecoin full-node
type API interface {
	ChainGetPath(context.Context, types.TipSetKey, types.TipSetKey) ([]*store.HeadChange, error)
	ChainGetGenesis(context.Context) (*types.TipSet, error)
}

// ChainSync provides methods to resolve chain syncing situations
type ChainSync struct {
	api API
}

// New returns a new ChainSync
func New(api API) *ChainSync {
	return &ChainSync{
		api: api,
	}
}

// Precedes returns true if from and to don't live in different chain forks, and
// from is at a lower epoch than to.
func (cs *ChainSync) Precedes(ctx context.Context, from, to types.TipSetKey) (bool, error) {
	fpath, err := cs.api.ChainGetPath(ctx, from, to)
	if err != nil {
		return false, err
	}
	if len(fpath) == 0 {
		return true, nil
	}
	norevert := fpath[0].Type == store.HCApply
	return norevert, nil
}

// ResolveBase returns the base TipSetKey that both left and right TipSetKey share,
// plus a Revert/Apply set of operations to get from last to new.
func ResolveBase(ctx context.Context, api API, left *types.TipSetKey, right types.TipSetKey) (*types.TipSetKey, []*types.TipSet, error) {
	var path []*types.TipSet
	if left == nil {
		genesis, err := api.ChainGetGenesis(ctx)
		if err != nil {
			return nil, nil, err
		}
		path = append(path, genesis)
		gtsk := types.NewTipSetKey(genesis.Cids()...)
		left = &gtsk
	}

	fpath, err := api.ChainGetPath(ctx, *left, right)
	if err != nil {
		return nil, nil, err
	}

	var base *types.TipSetKey
	for _, ts := range fpath {
		if ts.Type == store.HCApply {
			if base == nil {
				b := types.NewTipSetKey(ts.Val.Blocks()[0].Parents...)
				base = &b
			}
			path = append(path, ts.Val)
		}
	}
	return base, path, nil
}
