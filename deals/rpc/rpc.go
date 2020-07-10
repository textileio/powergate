package rpc

import (
	"context"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"

	logging "github.com/ipfs/go-log/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = logging.Logger("deals-rpc")

// Module is an implementation of the Deals Module.
type Module interface {
	Import(ctx context.Context, data io.Reader, isCAR bool) (cid.Cid, int64, error)
	Store(ctx context.Context, waddr string, dataCid cid.Cid, pieceSize uint64, dcfgs []deals.StorageDealConfig, minDuration uint64) ([]deals.StoreResult, error)
	Fetch(ctx context.Context, waddr string, cid cid.Cid) error
	Retrieve(ctx context.Context, waddr string, cid cid.Cid, CAREncoding bool) (io.ReadCloser, error)
	GetDealStatus(ctx context.Context, pcid cid.Cid) (storagemarket.StorageDealStatus, bool, error)
	Watch(ctx context.Context, proposals []cid.Cid) (<-chan deals.StorageDealInfo, error)
	ListStorageDealRecords(opts ...deals.ListDealRecordsOption) ([]deals.StorageDealRecord, error)
	ListRetrievalDealRecords(opts ...deals.ListDealRecordsOption) ([]deals.RetrievalDealRecord, error)
}

// RPC implements the gprc service.
type RPC struct {
	UnimplementedRPCServiceServer

	Module Module
}

type importResult struct {
	DataCid cid.Cid
	Size    int64
	Err     error
}

// New creates a new rpc service.
func New(dm Module) *RPC {
	return &RPC{
		Module: dm,
	}
}

// Import calls deals.Import.
func (s *RPC) Import(srv RPCService_ImportServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	var importParams *ImportParams
	switch payload := req.GetPayload().(type) {
	case *ImportRequest_ImportParams:
		importParams = payload.ImportParams
	default:
		return status.Errorf(codes.InvalidArgument, "expected ImportParams for ImportRequest.Payload but got %T", payload)
	}

	reader, writer := io.Pipe()

	importChannel := make(chan importResult)
	go importData(srv.Context(), s.Module, importParams, reader, importChannel)

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
		case *ImportRequest_Chunk:
			_, writeErr := writer.Write(payload.Chunk)
			if writeErr != nil {
				return writeErr
			}
		default:
			return status.Errorf(codes.InvalidArgument, "expected Chunk for ImportRequest.Payload but got %T", payload)
		}
	}

	importResult := <-importChannel
	if importResult.Err != nil {
		return importResult.Err
	}

	return srv.SendAndClose(&ImportResponse{DataCid: importResult.DataCid.String(), Size: importResult.Size})
}

func importData(ctx context.Context, dealsModule Module, importParams *ImportParams, r io.Reader, ch chan importResult) {
	defer close(ch)
	dcid, size, err := dealsModule.Import(ctx, r, importParams.IsCar)
	if err != nil {
		ch <- importResult{Err: err}
		return
	}
	ch <- importResult{DataCid: dcid, Size: size}
}

// Store calls deals.Store.
func (s *RPC) Store(ctx context.Context, req *StoreRequest) (*StoreResponse, error) {
	c, err := cid.Decode(req.DataCid)
	if err != nil {
		return nil, err
	}
	dealConfigs := make([]deals.StorageDealConfig, len(req.StorageDealConfigs))
	for i, dealConfig := range req.StorageDealConfigs {
		dealConfigs[i] = deals.StorageDealConfig{
			Miner:      dealConfig.Miner,
			EpochPrice: dealConfig.EpochPrice,
		}
	}
	results, err := s.Module.Store(ctx, req.Address, c, req.PieceSize, dealConfigs, req.MinDuration)
	if err != nil {
		return nil, err
	}
	ret := make([]*StoreResult, len(results))
	for i, result := range results {
		ret[i] = &StoreResult{
			ProposalCid: result.ProposalCid.String(),
			Message:     result.Message,
			Success:     result.Success,
			Config: &StorageDealConfig{
				Miner:      result.Config.Miner,
				EpochPrice: result.Config.EpochPrice,
			},
		}
	}
	return &StoreResponse{Results: ret}, nil
}

// Fetch calls deals.Fetch.
func (s *RPC) Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
	c, err := cid.Decode(req.Cid)
	if err != nil {
		return nil, err
	}
	if err := s.Module.Fetch(ctx, req.Address, c); err != nil {
		return nil, err
	}
	return &FetchResponse{}, nil
}

// Retrieve calls deals.Retreive.
func (s *RPC) Retrieve(req *RetrieveRequest, srv RPCService_RetrieveServer) error {
	c, err := cid.Parse(req.GetCid())
	if err != nil {
		return err
	}

	reader, err := s.Module.Retrieve(srv.Context(), req.GetAddress(), c, false)
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

// GetDealStatus calls deals.GetDealStatus.
func (s *RPC) GetDealStatus(ctx context.Context, req *GetDealStatusRequest) (*GetDealStatusResponse, error) {
	c, err := cid.Parse(req.ProposalCid)
	if err != nil {
		return nil, err
	}
	statusCode, slashed, err := s.Module.GetDealStatus(ctx, c)
	if err != nil {
		return nil, err
	}
	var status StorageDealStatus
	switch statusCode {
	case storagemarket.StorageDealProposalNotFound:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_NOT_FOUND
	case storagemarket.StorageDealProposalRejected:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_REJECTED
	case storagemarket.StorageDealProposalAccepted:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_ACCEPTED
	case storagemarket.StorageDealStaged:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_STAGED
	case storagemarket.StorageDealSealing:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_SEALING
	case storagemarket.StorageDealActive:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_ACTIVE
	case storagemarket.StorageDealFailing:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_FAILING
	case storagemarket.StorageDealNotFound:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_NOT_FOUND
	default:
		status = StorageDealStatus_STORAGE_DEAL_STATUS_UNSPECIFIED
	}
	return &GetDealStatusResponse{Status: status, Slashed: slashed}, nil
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
		dealInfo := &StorageDealInfo{
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

// ListStorageDealRecords calls deals.ListStorageDealRecords.
func (s *RPC) ListStorageDealRecords(ctx context.Context, req *ListStorageDealRecordsRequest) (*ListStorageDealRecordsResponse, error) {
	records, err := s.Module.ListStorageDealRecords(BuildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &ListStorageDealRecordsResponse{Records: ToRPCStorageDealRecords(records)}, nil
}

// ListRetrievalDealRecords calls deals.ListRetrievalDealRecords.
func (s *RPC) ListRetrievalDealRecords(ctx context.Context, req *ListRetrievalDealRecordsRequest) (*ListRetrievalDealRecordsResponse, error) {
	records, err := s.Module.ListRetrievalDealRecords(BuildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &ListRetrievalDealRecordsResponse{Records: ToRPCRetrievalDealRecords(records)}, nil
}

// BuildListDealRecordsOptions creates a ListDealRecordsOption from a ListDealRecordsConfig.
func BuildListDealRecordsOptions(conf *ListDealRecordsConfig) []deals.ListDealRecordsOption {
	var opts []deals.ListDealRecordsOption
	if conf != nil {
		opts = []deals.ListDealRecordsOption{
			deals.WithAscending(conf.Ascending),
			deals.WithDataCids(conf.DataCids...),
			deals.WithFromAddrs(conf.FromAddrs...),
			deals.WithIncludePending(conf.IncludePending),
			deals.WithIncludeFinal(conf.IncludeFinal),
		}
	}
	return opts
}

// ToRPCStorageDealRecords converts from Go to gRPC representations.
func ToRPCStorageDealRecords(records []deals.StorageDealRecord) []*StorageDealRecord {
	ret := make([]*StorageDealRecord, len(records))
	for i, r := range records {
		ret[i] = &StorageDealRecord{
			RootCid: r.RootCid.String(),
			Addr:    r.Addr,
			Time:    r.Time,
			Pending: r.Pending,
			DealInfo: &StorageDealInfo{
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

// ToRPCRetrievalDealRecords converts from Go to gRPC representations.
func ToRPCRetrievalDealRecords(records []deals.RetrievalDealRecord) []*RetrievalDealRecord {
	ret := make([]*RetrievalDealRecord, len(records))
	for i, r := range records {
		ret[i] = &RetrievalDealRecord{
			Addr: r.Addr,
			Time: r.Time,
			DealInfo: &RetrievalDealInfo{
				RootCid:                 r.DealInfo.RootCid.String(),
				Size:                    r.DealInfo.Size,
				MinPrice:                r.DealInfo.MinPrice,
				PaymentInterval:         r.DealInfo.PaymentInterval,
				PaymentIntervalIncrease: r.DealInfo.PaymentIntervalIncrease,
				Miner:                   r.DealInfo.Miner,
				MinerPeerId:             r.DealInfo.MinerPeerID,
			},
		}
	}
	return ret
}
