package api

import (
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/ffs"
)

type retrievalConfig struct {
	walletAddress string
	maxPrice      uint64
}

type Retrieval struct {
	ID            ffs.RetrievalID
	PayloadCid    cid.Cid
	PieceCid      cid.Cid
	Selector      string
	Miners        []string
	WalletAddress string
	MaxPrice      uint64
	JID           ffs.JobID
	CreatedAt     time.Time

	// Possibly empty
	RetrievalMiner string
	DataCid        cid.Cid
	Size           uint64
}

func (i *API) StartRetrieval(payloadCid, pieceCid cid.Cid, selector string, miners []string, opts ...PartialRetrievalOption) (Retrieval, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if !payloadCid.Defined() {
		return Retrieval{}, fmt.Errorf("payload cid is undefined")
	}
	if !pieceCid.Defined() {
		return Retrieval{}, fmt.Errorf("piece cid is undefined")
	}
	// ToDo: whenever text-based selectors are ready
	// we should try checking the prc.Selector is well-formated.

	defWalletAddress := i.cfg.DefaultStorageConfig.Cold.Filecoin.Addr
	rc := retrievalConfig{walletAddress: defWalletAddress}
	for _, o := range opts {
		o(&rc)
	}
	if !i.isManagedAddress(rc.walletAddress) {
		return Retrieval{}, fmt.Errorf("%s isn't a managed address", rc.walletAddress)
	}

	rID := ffs.NewRetrievalID()
	jid, err := i.sched.StartRetrieval(i.cfg.ID, rID, payloadCid, pieceCid, selector, miners, rc.walletAddress, rc.maxPrice)
	if err != nil {
		return Retrieval{}, fmt.Errorf("starting retrieval in scheduler: %s", err)
	}

	rr, err := i.is.putRetrievalRequest(rID, payloadCid, pieceCid, selector, miners, rc.walletAddress, rc.maxPrice, jid)
	if err != nil {
		return Retrieval{}, fmt.Errorf("saving new partial retrieval config: %s", err)
	}
	return retrievalRequestToRetrieval(rr), nil
}

func (i *API) GetRetrieval(prID ffs.RetrievalID) (Retrieval, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	rr, err := i.is.getRetrievalRequest(prID)
	if err == ErrNotFound {
		return Retrieval{}, err
	}
	if err != nil {
		return Retrieval{}, fmt.Errorf("getting retrieval request from store: %s", err)
	}
	r := retrievalRequestToRetrieval(rr)
	ri, err := i.sched.GetRetrievalInfo(r.ID)
	if err != nil {
		return Retrieval{}, fmt.Errorf("getting retrieval info: %s", err)
	}
	r.RetrievalMiner = ri.RetrievalMiner
	r.DataCid = ri.DataCid

	return r, nil

}

func (i *API) RemoveRetrieval(partialCid cid.Cid) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	// ToDo:
	// Check retrieval exists
	// Check retrieval ID was done by this API
	// Remove from scheduler (reference counting)
	// Remove from instance-store

	return nil
}

func retrievalRequestToRetrieval(rr retrievalRequest) Retrieval {
	return Retrieval{
		ID:            rr.ID,
		PayloadCid:    rr.PayloadCid,
		PieceCid:      rr.PieceCid,
		Selector:      rr.Selector,
		Miners:        rr.Miners,
		WalletAddress: rr.WalletAddress,
		MaxPrice:      rr.MaxPrice,
		JID:           rr.JID,
		CreatedAt:     rr.CreatedAt,
	}
}
