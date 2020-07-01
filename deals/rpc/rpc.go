package rpc

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"

	logging "github.com/ipfs/go-log/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = logging.Logger("deals-rpc")

// RPC implements the gprc service.
type RPC struct {
	UnimplementedRPCServiceServer

	Module *deals.Module
}

type storeResult struct {
	DataCid      cid.Cid
	ProposalCids []cid.Cid
	FailedDeals  []deals.StorageDealConfig
	Err          error
}

// New creates a new rpc service.
func New(dm *deals.Module) *RPC {
	return &RPC{
		Module: dm,
	}
}

func store(ctx context.Context, dealsModule *deals.Module, storeParams *StoreParams, r io.Reader, ch chan storeResult) {
	defer close(ch)
	dealConfigs := make([]deals.StorageDealConfig, len(storeParams.GetDealConfigs()))
	for i, dealConfig := range storeParams.GetDealConfigs() {
		dealConfigs[i] = deals.StorageDealConfig{
			Miner:      dealConfig.GetMiner(),
			EpochPrice: dealConfig.GetEpochPrice(),
		}
	}
	dcid, size, err := dealsModule.Import(ctx, r, false)
	if err != nil {
		ch <- storeResult{Err: err}
		return
	}
	sr, err := dealsModule.Store(ctx, storeParams.GetAddress(), dcid, uint64(size), dealConfigs, storeParams.GetMinDuration())
	if err != nil {
		ch <- storeResult{Err: err}
		return
	}
	var failedDeals []deals.StorageDealConfig
	var pcids []cid.Cid
	for _, res := range sr {
		if res.Success {
			pcids = append(pcids, res.ProposalCid)
		} else {
			failedDeals = append(failedDeals, res.Config)
		}
	}
	ch <- storeResult{DataCid: dcid, ProposalCids: pcids, FailedDeals: failedDeals}
}

// Store calls deals.Store.
func (s *RPC) Store(srv RPCService_StoreServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	var storeParams *StoreParams
	switch payload := req.GetPayload().(type) {
	case *StoreRequest_StoreParams:
		storeParams = payload.StoreParams
	default:
		return status.Errorf(codes.InvalidArgument, "expected StoreParams for StoreRequest.Payload but got %T", payload)
	}

	reader, writer := io.Pipe()

	storeChannel := make(chan storeResult)
	go store(srv.Context(), s.Module, storeParams, reader, storeChannel)

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			_ = writer.Close()
			break
		} else if err != nil {
			_ = writer.CloseWithError(err)
			break
		}
		switch payload := req.GetPayload().(type) {
		case *StoreRequest_Chunk:
			_, writeErr := writer.Write(payload.Chunk)
			if writeErr != nil {
				return writeErr
			}
		default:
			return status.Errorf(codes.InvalidArgument, "expected Chunk for StoreRequest.Payload but got %T", payload)
		}
	}

	storeResult := <-storeChannel
	if storeResult.Err != nil {
		return storeResult.Err
	}

	replyCids := make([]string, len(storeResult.ProposalCids))
	for i, cid := range storeResult.ProposalCids {
		replyCids[i] = cid.String()
	}

	replyFailedDeals := make([]*DealConfig, len(storeResult.FailedDeals))
	for i, dealConfig := range storeResult.FailedDeals {
		replyFailedDeals[i] = &DealConfig{Miner: dealConfig.Miner, EpochPrice: dealConfig.EpochPrice}
	}

	return srv.SendAndClose(&StoreResponse{DataCid: storeResult.DataCid.String(), ProposalCids: replyCids, FailedDeals: replyFailedDeals})
}

// Watch calls deals.Watch.
func (s *RPC) Watch(req *WatchRequest, srv RPCService_WatchServer) error {
	proposals := make([]cid.Cid, len(req.GetProposals()))
	for i, proposal := range req.GetProposals() {
		id, err := cid.Decode(proposal)
		if err != nil {
			return err
		}
		proposals[i] = id
	}
	ch, err := s.Module.Watch(srv.Context(), proposals)
	if err != nil {
		return err
	}

	for update := range ch {
		var strProposalCid string
		if update.ProposalCid.Defined() {
			strProposalCid = update.ProposalCid.String()
		}
		dealInfo := &DealInfo{
			ProposalCid:   strProposalCid,
			StateId:       update.StateID,
			StateName:     update.StateName,
			Miner:         update.Miner,
			PieceCid:      update.PieceCID.String(),
			Size:          update.Size,
			PricePerEpoch: update.PricePerEpoch,
			Duration:      update.Duration,
		}
		if err := srv.Send(&WatchResponse{DealInfo: dealInfo}); err != nil {
			log.Errorf("sending response: %v", err)
			return err
		}
	}
	return nil
}

// Retrieve calls deals.Retreive.
func (s *RPC) Retrieve(req *RetrieveRequest, srv RPCService_RetrieveServer) error {
	cid, err := cid.Parse(req.GetCid())
	if err != nil {
		return err
	}

	reader, err := s.Module.Retrieve(srv.Context(), req.GetAddress(), cid, false)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Errorf("closing reader on Retrieve: %s", err)
		}
	}()

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if sendErr := srv.Send(&RetrieveResponse{Chunk: buffer[:bytesRead]}); sendErr != nil {
			return sendErr
		}
		if err == io.EOF {
			return nil
		}
	}
}

// FinalDealRecords calls deals.FinalDealRecords.
func (s *RPC) FinalDealRecords(ctx context.Context, req *FinalDealRecordsRequest) (*FinalDealRecordsResponse, error) {
	records, err := s.Module.FinalDealRecords()
	if err != nil {
		return nil, err
	}
	return &FinalDealRecordsResponse{Records: toRPCDealRecords(records)}, nil
}

// PendingDealRecords calls deals.PendingDealRecords.
func (s *RPC) PendingDealRecords(ctx context.Context, req *PendingDealRecordsRequest) (*PendingDealRecordsResponse, error) {
	records, err := s.Module.PendingDealRecords()
	if err != nil {
		return nil, err
	}
	return &PendingDealRecordsResponse{Records: toRPCDealRecords(records)}, nil
}

// AllDealRecords calls deals.AllDealRecords.
func (s *RPC) AllDealRecords(ctx context.Context, req *AllDealRecordsRequest) (*AllDealRecordsResponse, error) {
	records, err := s.Module.AllDealRecords()
	if err != nil {
		return nil, err
	}
	return &AllDealRecordsResponse{Records: toRPCDealRecords(records)}, nil
}

// RetrievalRecords calls deals.RetrievalRecords.
func (s *RPC) RetrievalRecords(ctx context.Context, req *RetrievalRecordsRequest) (*RetrievalRecordsResponse, error) {
	records, err := s.Module.RetrievalRecords()
	if err != nil {
		return nil, err
	}
	ret := make([]*RetrievalRecord, len(records))
	for i, r := range records {
		ret[i] = &RetrievalRecord{
			Addr: r.Addr,
			Time: r.Time,
			RetrievalInfo: &RetrievalInfo{
				PieceCid:                r.RetrievalInfo.PieceCID.String(),
				Size:                    r.RetrievalInfo.Size,
				MinPrice:                r.RetrievalInfo.MinPrice,
				PaymentInterval:         r.RetrievalInfo.PaymentInterval,
				PaymentIntervalIncrease: r.RetrievalInfo.PaymentIntervalIncrease,
				Miner:                   r.RetrievalInfo.Miner,
				MinerPeerId:             r.RetrievalInfo.MinerPeerID,
			},
		}
	}
	return &RetrievalRecordsResponse{Records: ret}, nil
}

func toRPCDealRecords(records []deals.DealRecord) []*DealRecord {
	ret := make([]*DealRecord, len(records))
	for i, r := range records {
		ret[i] = &DealRecord{
			Addr:    r.Addr,
			Time:    r.Time,
			Pending: r.Pending,
			DealInfo: &DealInfo{
				ProposalCid:     r.DealInfo.ProposalCid.String(),
				StateId:         r.DealInfo.StateID,
				StateName:       r.DealInfo.StateName,
				Miner:           r.DealInfo.Miner,
				PieceCid:        r.DealInfo.PieceCID.String(),
				Size:            r.DealInfo.Size,
				PricePerEpoch:   r.DealInfo.PricePerEpoch,
				StartEpoch:      r.DealInfo.StartEpoch,
				Duration:        r.DealInfo.Duration,
				DealId:          r.DealInfo.DealID,
				ActivationEpoch: r.DealInfo.ActivationEpoch,
				Msg:             r.DealInfo.Message,
			},
		}
	}
	return ret
}
