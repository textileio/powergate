package chainsync

import (
	"context"
	"fmt"

	"github.com/textileio/filecoin/lotus/types"
)

// ToDo: consider avoiding the API and ask for deps in parameters

// API provides an abstraction to a lotus node
type API interface {
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainHead(context.Context) (*types.TipSet, error)
}

// GetPathToHead returns an ordered tipset slice from the base tipset to the current
// head.
func GetPathToHead(ctx context.Context, c API, base types.TipSetKey) ([]*types.TipSet, error) {
	// ToDo: this is a temporary impl to start preparing for a possible new API that can
	// return HeadChanges from a checkpoint tipset to the head.
	bts, err := c.ChainGetTipSet(ctx, base)
	if err != nil {
		return nil, err
	}
	ts, err := c.ChainHead(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when getting heaviest head from chain: %s", err)
	}
	if ts.Height < bts.Height {
		return nil, fmt.Errorf("reversed node or chain rollback, current height %d, last height %d", ts.Height, bts.Height)
	}
	path := make([]*types.TipSet, 0, ts.Height-bts.Height+1)
	for ts.Height >= bts.Height {
		path = append(path, ts)
		ts, err = c.ChainGetTipSet(ctx, types.NewTipSetKey(ts.Blocks[0].Parents...))
		if err != nil {
			return nil, err
		}
	}
	return path, nil
}
