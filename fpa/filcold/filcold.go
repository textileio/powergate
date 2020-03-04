package filcold

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-car"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logger "github.com/ipfs/go-log/v2"
	"github.com/textileio/fil-tools/deals"
	"github.com/textileio/fil-tools/fpa"
)

var (
	log = logger.Logger("filcold")
)

type FilCold struct {
	ms  fpa.MinerSelector
	dm  *deals.Module
	dag format.DAGService
}

var _ fpa.ColdLayer = (*FilCold)(nil)

func New(ms fpa.MinerSelector, dm *deals.Module, dag format.DAGService) *FilCold {
	return &FilCold{
		ms:  ms,
		dm:  dm,
		dag: dag,
	}
}

func (fc *FilCold) Store(ctx context.Context, c cid.Cid, conf fpa.ColdConfig) (fpa.ColdInfo, error) {
	var ci fpa.ColdInfo
	ci.Filecoin = fpa.FilInfo{
		Duration: uint64(1000), // ToDo: will evolve as instance/methdo-call has more config. Use ColdConfig
	}

	config, err := makeStorageConfig(ctx, fc.ms) // ToDo: will evolve as instance/methdo-call has more config
	if err != nil {
		return ci, fmt.Errorf("selecting miners to make the deal: %s", err)
	}
	r := ipldToFileTransform(ctx, fc.dag, c)
	log.Infof("storing deals in filecoin...")
	dataCid, result, err := fc.dm.Store(ctx, conf.Filecoin.WalletAddr, r, config, ci.Filecoin.Duration)
	if err != nil {
		return ci, err
	}
	ci.Filecoin.PayloadCID = dataCid

	notDone := make(map[cid.Cid]struct{})
	propcids := make([]cid.Cid, len(result))
	ci.Filecoin.Proposals = make([]fpa.FilStorage, len(result))
	for i, d := range result {
		ci.Filecoin.Proposals[i] = fpa.FilStorage{
			ProposalCid: d.ProposalCid,
			Failed:      !d.Success,
		}
		propcids[i] = d.ProposalCid
		notDone[d.ProposalCid] = struct{}{}
	}

	log.Infof("watching deals unfold...")
	chDi, err := fc.dm.Watch(ctx, propcids)

	for di := range chDi {
		if di.StateID == storagemarket.StorageDealActive {
			delete(notDone, di.ProposalCid)
		}
		// ToDo: handle other states
		if len(notDone) == 0 {
			break
		}
	}
	log.Infof("done")
	return ci, nil
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

func makeStorageConfig(ctx context.Context, ms fpa.MinerSelector) ([]deals.StorageDealConfig, error) {
	mps, err := ms.GetTopMiners(1) // ToDo: hardcoded 1 will change when config will be added to instance/method-call
	if err != nil {
		return nil, err
	}
	res := make([]deals.StorageDealConfig, len(mps))
	for i, m := range mps {
		res[i] = deals.StorageDealConfig{Miner: m.Addr, EpochPrice: m.EpochPrice}
	}
	return res, nil
}
