package reptop

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/chain/types"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/powergate/ffs"
	askRunner "github.com/textileio/powergate/index/ask/runner"
	"github.com/textileio/powergate/lotus"
	"github.com/textileio/powergate/reputation"
)

var (
	log = logger.Logger("reptop")
)

// RepTop is a ffs.MinerSelector implementation that returns the top N
// miners from a Reputations Module and an Ask Index.
type RepTop struct {
	rm *reputation.Module
	ai *askRunner.Runner
	cb lotus.ClientBuilder
}

var _ ffs.MinerSelector = (*RepTop)(nil)

// New returns a new RetTop instance that uses the specified Reputation Module
// to select miners and the AskIndex for their epoch prices.
func New(rm *reputation.Module, ai *askRunner.Runner) *RepTop {
	return &RepTop{
		rm: rm,
		ai: ai,
	}
}

// GetMiners returns n miners using the configured Reputation Module and
// Ask Index.
func (rt *RepTop) GetMiners(n int, f ffs.MinerSelectorFilter) ([]ffs.MinerProposal, error) {
	if n < 1 {
		return nil, fmt.Errorf("the number of miners should be greater than zero")
	}

	// We start directly targeting trusted miners without relying on reputation index.
	// This is done to circumvent index building restrictions such as excluding miners
	// with zero power. Trusted miners are trusted by definition, so if they have zero
	// power, that's a risk that the client is accepting.
	trustedMiners := rt.genTrustedMiners(f)

	// The remaining needed miners are gathered from the reputation index.
	reputationMiners, err := rt.genFromReputation(f, n-len(trustedMiners))
	if err != nil {
		return nil, fmt.Errorf("getting miners from reputation module: %s", err)
	}

	// Return both lists.
	return append(trustedMiners, reputationMiners...), nil
}

func (rt *RepTop) genTrustedMiners(f ffs.MinerSelectorFilter) []ffs.MinerProposal {
	ret := make([]ffs.MinerProposal, 0, len(f.TrustedMiners))
	for _, m := range f.TrustedMiners {
		mp, err := rt.getMinerProposal(f, m)
		if err != nil {
			log.Warnf("trusted miner query asking: %s", err)
			continue
		}
		ret = append(ret, mp)
	}

	return ret
}

func (rt *RepTop) genFromReputation(f ffs.MinerSelectorFilter, n int) ([]ffs.MinerProposal, error) {
	ms, err := rt.rm.QueryMiners(f.TrustedMiners, f.CountryCodes, nil)
	if err != nil {
		return nil, fmt.Errorf("getting miners from reputation module: %s", err)
	}
	if len(ms) < n {
		return nil, fmt.Errorf("not enough miners from reputation module to satisfy the constraints")
	}

	res := make([]ffs.MinerProposal, 0, n)
	for _, m := range ms {
		mp, err := rt.getMinerProposal(f, m.Addr)
		if err != nil {
			continue // Cached miner isn't replying to query-ask now, skip.
		}
		res = append(res, mp)
		if len(res) == n {
			break
		}
	}
	if len(res) < n {
		return nil, fmt.Errorf("not enough miners satisfy the miner selector constraints")
	}
	return res, nil
}

func (rt *RepTop) getMinerProposal(f ffs.MinerSelectorFilter, addrStr string) (ffs.MinerProposal, error) {
	c, cls, err := rt.cb(context.Background())
	if err != nil {
		return ffs.MinerProposal{}, fmt.Errorf("creating lotus client: %s", err)
	}
	defer cls()

	addr, err := address.NewFromString(addrStr)
	if err != nil {
		return ffs.MinerProposal{}, fmt.Errorf("miner address is invalid: %s", err)
	}
	ctx, cls := context.WithTimeout(context.Background(), time.Second*10)
	defer cls()

	mi, err := c.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return ffs.MinerProposal{}, fmt.Errorf("getting miner %s info: %s", addr, err)
	}

	type chAskRes struct {
		Error string
		Ask   *storagemarket.StorageAsk
	}
	chAsk := make(chan chAskRes)
	go func() {
		sask, err := c.ClientQueryAsk(ctx, *mi.PeerId, addr)
		if err != nil {
			chAsk <- chAskRes{Error: err.Error()}
			return
		}

		chAsk <- chAskRes{Ask: sask}
	}()

	select {
	case <-time.After(time.Second * 10):
		return ffs.MinerProposal{}, fmt.Errorf("query asking timed out")
	case r := <-chAsk:
		if r.Error != "" {
			return ffs.MinerProposal{}, fmt.Errorf("query ask had controlled error: %s", r.Error)
		}
		if f.MaxPrice > 0 && r.Ask.Price.Uint64() > f.MaxPrice {
			return ffs.MinerProposal{}, fmt.Errorf("miner doesn't satisfy price constraints %d>%d", r.Ask.Price, f.MaxPrice)
		}
		if f.PieceSize < uint64(r.Ask.MinPieceSize) || f.PieceSize > uint64(r.Ask.MaxPieceSize) {
			return ffs.MinerProposal{}, fmt.Errorf("miner doesn't satisfy piece size constraints: %d<%d<%d", r.Ask.MinPieceSize, f.PieceSize, r.Ask.MaxPieceSize)
		}

		return ffs.MinerProposal{Addr: addrStr, EpochPrice: r.Ask.Price.Uint64()}, nil
	}
}
