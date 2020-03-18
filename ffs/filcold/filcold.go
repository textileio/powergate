package filcold

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logger "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-car"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
)

var (
	log = logger.Logger("ffs-filcold")
)

// FilCold is a ffs.ColdStorage implementation that stores data in Filecoin.
type FilCold struct {
	ms    ffs.MinerSelector
	dm    *deals.Module
	dag   format.DAGService
	chain FilChain
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// FilChain is an abstraction of a Filecoin node to get information of the network.
type FilChain interface {
	GetHeight(context.Context) (uint64, error)
}

// New returns a new FilCold instance
func New(ms ffs.MinerSelector, dm *deals.Module, dag format.DAGService, chain FilChain) *FilCold {
	return &FilCold{
		ms:    ms,
		dm:    dm,
		dag:   dag,
		chain: chain,
	}
}

// Retrieve returns the original data Cid, from the CAR encoded data Cid. The returned Cid is available in the
// car.Store received as a parameter.
func (fc *FilCold) Retrieve(ctx context.Context, dataCid cid.Cid, cs car.Store, waddr string) (cid.Cid, error) {
	carR, err := fc.dm.Retrieve(ctx, waddr, dataCid)
	if err != nil {
		return cid.Undef, fmt.Errorf("retrieving from deal module: %s", err)
	}
	defer carR.Close()
	h, err := car.LoadCar(cs, carR)
	if err != nil {
		return cid.Undef, fmt.Errorf("loading car to carstore: %s", err)
	}
	if len(h.Roots) != 1 {
		return cid.Undef, fmt.Errorf("car header doesn't have a single root: %d", len(h.Roots))
	}
	return h.Roots[0], nil
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation.
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, waddr string, fcfg ffs.FilConfig) (ffs.FilInfo, error) {
	f := ffs.MinerSelectorFilter{
		Blacklist:    fcfg.Blacklist,
		CountryCodes: fcfg.CountryCodes,
	}
	cfgs, err := makeDealConfigs(ctx, fc.ms, fcfg.RepFactor, f)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("making deal configs: %s", err)
	}

	cid, props, err := fc.makeDeals(ctx, c, cfgs, waddr, fcfg)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("executing deals: %s", err)
	}

	return ffs.FilInfo{
		DataCid:   cid,
		Proposals: props,
	}, nil
}

// EnsureRenewals analyzes a FilInfo state for a Cid and executes renweals considering the FilConfig desired configuration.
func (fc *FilCold) EnsureRenewals(ctx context.Context, c cid.Cid, inf ffs.FilInfo, waddr string, fcfg ffs.FilConfig) (ffs.FilInfo, error) {
	var activeMiners []string
	for _, p := range inf.Proposals {
		if p.Active {
			activeMiners = append(activeMiners, p.Miner)
		}
	}
	height, err := fc.chain.GetHeight(ctx)
	if err != nil {
		return ffs.FilInfo{}, fmt.Errorf("get current filecoin height: %s", err)
	}
	for i, p := range inf.Proposals {
		if !p.Active || p.Renewed {
			continue
		}
		expiry := p.ActivationEpoch + uint64(p.Duration)
		renewalHeight := expiry - uint64(fcfg.Renew.Threshold)
		if renewalHeight <= height {
			newProposal, err := fc.renewDeal(ctx, c, waddr, p, activeMiners, fcfg)
			if err != nil {
				log.Errorf("renewing deal %s: %s", p.ProposalCid, err)
				continue
			}
			inf.Proposals = append(inf.Proposals, newProposal)
			inf.Proposals[i].Renewed = true
		}
	}
	return inf, nil
}

func (fc *FilCold) renewDeal(ctx context.Context, c cid.Cid, waddr string, p ffs.FilStorage, activeMiners []string, fcfg ffs.FilConfig) (ffs.FilStorage, error) {
	f := ffs.MinerSelectorFilter{
		Blacklist: activeMiners,
	}
	dealConfig, err := makeDealConfigs(ctx, fc.ms, 1, f)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("making new deal config: %s", err)
	}

	_, props, err := fc.makeDeals(ctx, c, dealConfig, waddr, fcfg)
	if err != nil {
		return ffs.FilStorage{}, fmt.Errorf("executing renewed deal: %s", err)
	}
	if len(props) != 1 {
		return ffs.FilStorage{}, fmt.Errorf("unsuccessful renewal")
	}
	return props[0], nil
}

func (fc *FilCold) makeDeals(ctx context.Context, c cid.Cid, cfgs []deals.StorageDealConfig, waddr string, fcfg ffs.FilConfig) (cid.Cid, []ffs.FilStorage, error) {
	r := ipldToFileTransform(ctx, fc.dag, c)

	var sres []deals.StoreResult
	dataCid, sres, err := fc.dm.Store(ctx, waddr, r, cfgs, uint64(fcfg.DealDuration))
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("storing deals in deal module: %s", err)
	}

	proposals, err := fc.waitForDeals(ctx, sres, fcfg.DealDuration)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("waiting for deals to finish: %s", err)
	}
	return dataCid, proposals, nil
}

func (fc *FilCold) waitForDeals(ctx context.Context, storeResults []deals.StoreResult, duration int64) ([]ffs.FilStorage, error) {
	notDone := make(map[cid.Cid]struct{})
	var inProgressDeals []cid.Cid
	proposals := make(map[cid.Cid]*ffs.FilStorage)
	for _, d := range storeResults {
		if !d.Success {
			log.Warnf("failed store result")
			continue
		}
		proposals[d.ProposalCid] = &ffs.FilStorage{
			ProposalCid: d.ProposalCid,
			Active:      true,
			Duration:    duration,
			Miner:       d.Config.Miner,
		}
		inProgressDeals = append(inProgressDeals, d.ProposalCid)
		notDone[d.ProposalCid] = struct{}{}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chDi, err := fc.dm.Watch(ctx, inProgressDeals)
	if err != nil {
		return nil, err
	}
	for di := range chDi {
		log.Infof("watching pending %d deals unfold...", len(notDone))
		fs, ok := proposals[di.ProposalCid]
		if !ok {
			continue
		}
		if di.StateID == storagemarket.DealComplete {
			fs.ActivationEpoch = di.ActivationEpoch
			delete(notDone, di.ProposalCid)
		} else if di.StateID == storagemarket.DealError || di.StateID == storagemarket.DealFailed {
			delete(proposals, di.ProposalCid)
			delete(notDone, di.ProposalCid)
		}
		if len(notDone) == 0 {
			break
		}
	}
	log.Infof("deals reached final state")

	res := make([]ffs.FilStorage, 0, len(proposals))
	for _, v := range proposals {
		res = append(res, *v)
	}
	return res, nil
}

func ipldToFileTransform(ctx context.Context, dag format.DAGService, c cid.Cid) io.Reader {
	r, w := io.Pipe()
	go func() {
		if err := car.WriteCar(ctx, dag, []cid.Cid{c}, w); err != nil {
			w.CloseWithError(err)
		}
		w.Close()
	}()
	return r
}

func makeDealConfigs(ctx context.Context, ms ffs.MinerSelector, cantMiners int, f ffs.MinerSelectorFilter) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetMiners(cantMiners, f)
	if err != nil {
		return nil, fmt.Errorf("getting miners from minerselector: %s", err)
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice}
	}
	return res, nil
}
