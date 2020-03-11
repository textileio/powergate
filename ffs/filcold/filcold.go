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
	ms  ffs.MinerSelector
	dm  *deals.Module
	dag format.DAGService
}

var _ ffs.ColdStorage = (*FilCold)(nil)

// New returns a new FilCold instance
func New(ms ffs.MinerSelector, dm *deals.Module, dag format.DAGService) *FilCold {
	return &FilCold{
		ms:  ms,
		dm:  dm,
		dag: dag,
	}
}

// Store stores a Cid in Filecoin considering the configuration provided. The Cid is retrieved using
// the DAGService registered on instance creation. Currently, a default configuration is used.
// (TODO: ColdConfig will enable more configurations in the future)
func (fc *FilCold) Store(ctx context.Context, c cid.Cid, waddr string, conf ffs.ColdConfig) (ffs.ColdInfo, error) {
	var ci ffs.ColdInfo
	config, err := makeStorageConfig(ctx, fc.ms, conf.Filecoin)
	if err != nil {
		return ci, fmt.Errorf("selecting miners to make the deal: %s", err)
	}
	r := ipldToFileTransform(ctx, fc.dag, c)

	log.Infof("storing deals in filecoin...")
	var sres []deals.StoreResult
	ci.Filecoin.PayloadCID, sres, err = fc.dm.Store(ctx, waddr, r, config, uint64(conf.Filecoin.DealDuration))
	if err != nil {
		return ci, fmt.Errorf("storing deals in deal manager: %s", err)
	}

	if ci.Filecoin.Proposals, err = fc.waitForDeals(ctx, sres, conf.Filecoin.DealDuration); err != nil {
		return ci, fmt.Errorf("waiting for deals to finish: %s", err)
	}
	return ci, nil
}

func (fc *FilCold) waitForDeals(ctx context.Context, storeResults []deals.StoreResult, duration int64) ([]ffs.FilStorage, error) {
	notDone := make(map[cid.Cid]struct{})
	var inProgressDeals []cid.Cid
	proposals := make(map[cid.Cid]*ffs.FilStorage)
	for _, d := range storeResults {
		proposals[d.ProposalCid] = &ffs.FilStorage{
			ProposalCid: d.ProposalCid,
			Failed:      !d.Success,
			Duration:    duration,
			Miner:       d.Config.Miner,
		}
		if d.Success {
			inProgressDeals = append(inProgressDeals, d.ProposalCid)
			notDone[d.ProposalCid] = struct{}{}
		}
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
			fs.Failed = true
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

func makeStorageConfig(ctx context.Context, ms ffs.MinerSelector, conf ffs.FilecoinConfig) ([]deals.StorageDealConfig, error) {
	filters := ffs.MinerSelectorFilter{
		Blacklist:    conf.Blacklist,
		CountryCodes: conf.CountryCodes,
	}
	mps, err := ms.GetMiners(conf.RepFactor, filters)
	if err != nil {
		return nil, err
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice}
	}
	return res, nil
}
