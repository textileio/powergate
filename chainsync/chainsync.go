package chainsync

import (
	"context"
	"fmt"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/powergate/v2/lotus"
)

const (
	hcApply = "apply"
	// For completeness:
	// hcRevert = "revert".
	// hcCurrent = "current".
)

// ChainSync provides methods to resolve chain syncing situations.
type ChainSync struct {
	clientBuilder lotus.ClientBuilder
}

// New returns a new ChainSync.
func New(clientBuilder lotus.ClientBuilder) *ChainSync {
	return &ChainSync{
		clientBuilder: clientBuilder,
	}
}

// Precedes returns true if from and to don't live in different chain forks, and
// from is at a lower epoch than to.
func (cs *ChainSync) Precedes(ctx context.Context, from, to types.TipSetKey) (bool, error) {
	client, cls, err := cs.clientBuilder(ctx)
	if err != nil {
		return false, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()
	fpath, err := client.ChainGetPath(ctx, from, to)
	if err != nil {
		return false, fmt.Errorf("getting path from %v to %v: %s", from.Cids(), to.Cids(), err)
	}
	if len(fpath) == 0 {
		return true, nil
	}
	norevert := fpath[0].Type == hcApply
	return norevert, nil
}

// ResolveBase returns the base TipSetKey that both left and right TipSetKey share,
// plus a Revert/Apply set of operations to get from last to new.
func ResolveBase(ctx context.Context, api *api.FullNodeStruct, left *types.TipSetKey, right types.TipSetKey) (*types.TipSetKey, []*types.TipSet, error) {
	var path []*types.TipSet
	if left == nil {
		genesis, err := api.ChainGetGenesis(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("getting genesis tipset: %s", err)
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
		if ts.Type == hcApply {
			if base == nil {
				b := types.NewTipSetKey(ts.Val.Blocks()[0].Parents...)
				base = &b
			}
			path = append(path, ts.Val)
		}
	}
	return base, path, nil
}
