package client

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
	cid "github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/deals/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deals provides an API for managing deals and storing data.
type Deals struct {
	client rpc.RPCServiceClient
}

// WatchEvent is used to send data or error values for Watch.
type WatchEvent struct {
	Deal deals.StorageDealInfo
	Err  error
}

// ListDealRecordsOption updates a ListDealRecordsConfig.
type ListDealRecordsOption func(*rpc.ListDealRecordsConfig)

// WithFromAddrs limits the results deals initiated from the provided wallet addresses.
// If WithDataCids is also provided, this is an AND operation.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
// If WithFromAddrs is also provided, this is an AND operation.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithIncludeFinal specifies whether or not to include final deals in the results. Default is false.
// Ignored for ListRetrievalDealRecords.
func WithIncludeFinal(includeFinal bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.IncludeFinal = includeFinal
	}
}

// WithAscending specifies to sort the results in ascending order. Default is descending order.
// Records are sorted by timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}

// Import imports raw data in the Filecoin client. The isCAR flag indicates if the data
// is already in CAR format, so it shouldn't be encoded into a UnixFS DAG in the Filecoin client.
// It returns the imported data cid and the data size.
func (d *Deals) Import(ctx context.Context, data io.Reader, isCAR bool) (cid.Cid, int64, error) {
	stream, err := d.client.Import(ctx)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("calling Import: %v", err)
	}

	importParams := &rpc.ImportParams{
		IsCar: isCAR,
	}
	innerReq := &rpc.ImportRequest_ImportParams{ImportParams: importParams}

	if err = stream.Send(&rpc.ImportRequest{Payload: innerReq}); err != nil {
		return cid.Undef, 0, fmt.Errorf("calling Send: %v", err)
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return cid.Undef, 0, fmt.Errorf("reading buffer: %v", err)
		}
		sendErr := stream.Send(&rpc.ImportRequest{Payload: &rpc.ImportRequest_Chunk{Chunk: buffer[:bytesRead]}})
		if sendErr != nil {
			return cid.Undef, 0, fmt.Errorf("calling Send: %v", err)
		}
		if err == io.EOF {
			break
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("calling CloseAndRecv: %v", err)
	}

	c, err := cid.Decode(res.DataCid)
	if err != nil {
		return cid.Undef, 0, fmt.Errorf("decoding cid: %v", err)
	}

	return c, res.Size, nil
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs.
func (d *Deals) Store(ctx context.Context, waddr string, dataCid cid.Cid, pieceSize uint64, dcfgs []deals.StorageDealConfig, minDuration uint64) ([]deals.StoreResult, error) {
	dealConfigs := make([]*rpc.StorageDealConfig, len(dcfgs))
	for i, dealConfig := range dcfgs {
		dealConfigs[i] = &rpc.StorageDealConfig{
			EpochPrice: dealConfig.EpochPrice,
			Miner:      dealConfig.Miner,
		}
	}
	req := &rpc.StoreRequest{
		Address:            waddr,
		DataCid:            dataCid.String(),
		PieceSize:          pieceSize,
		StorageDealConfigs: dealConfigs,
		MinDuration:        minDuration,
	}
	res, err := d.client.Store(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("calling Store: %v", err)
	}
	results := make([]deals.StoreResult, len(res.Results))
	for i, result := range res.Results {
		c, err := cid.Decode(result.ProposalCid)
		if err != nil {
			return nil, fmt.Errorf("decoding cid: %v", err)
		}
		results[i] = deals.StoreResult{
			ProposalCid: c,
			Success:     result.Success,
			Message:     result.Message,
			Config: deals.StorageDealConfig{
				Miner:      result.Config.Miner,
				EpochPrice: result.Config.EpochPrice,
			},
		}
	}
	return results, nil
}

// Fetch fetches deal data to the underlying blockstore of the Filecoin client.
// This API is meant for clients that use external implementations of blockstores with
// their own API, e.g: IPFS.
func (d *Deals) Fetch(ctx context.Context, waddr string, cid cid.Cid) error {
	req := &rpc.FetchRequest{
		Address: waddr,
		Cid:     cid.String(),
	}
	if _, err := d.client.Fetch(ctx, req); err != nil {
		return fmt.Errorf("calling Fetch: %v", err)
	}
	return nil
}

// Retrieve is used to fetch data from filecoin.
func (d *Deals) Retrieve(ctx context.Context, waddr string, cid cid.Cid, CAREncoding bool) (io.ReadCloser, error) {
	req := &rpc.RetrieveRequest{
		Address:     waddr,
		Cid:         cid.String(),
		CarEncoding: CAREncoding,
	}
	stream, err := d.client.Retrieve(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("calling Retrieve: %v", err)
	}

	reader, writer := io.Pipe()

	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				_ = writer.Close()
				break
			} else if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
			_, err = writer.Write(reply.GetChunk())
			if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
		}
	}()

	return reader, nil
}

// GetDealStatus returns the current status of the deal, and a flag indicating if the miner of the deal was slashed.
// If the deal doesn't exist, *or has expired* it will return ErrDealNotFound. There's not actual way of distinguishing
// both scenarios in Lotus.
func (d *Deals) GetDealStatus(ctx context.Context, pcid cid.Cid) (storagemarket.StorageDealStatus, bool, error) {
	res, err := d.client.GetDealStatus(ctx, &rpc.GetDealStatusRequest{ProposalCid: pcid.String()})
	if err != nil {
		return storagemarket.StorageDealUnknown, false, fmt.Errorf("calling GetDealStatus: %v", err)
	}
	var status storagemarket.StorageDealStatus
	switch res.Status {
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_UNSPECIFIED:
		status = storagemarket.StorageDealUnknown
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_NOT_FOUND:
		status = storagemarket.StorageDealProposalNotFound
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_REJECTED:
		status = storagemarket.StorageDealProposalRejected
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_PROPOSAL_ACCEPTED:
		status = storagemarket.StorageDealProposalAccepted
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_STAGED:
		status = storagemarket.StorageDealStaged
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_SEALING:
		status = storagemarket.StorageDealSealing
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_ACTIVE:
		status = storagemarket.StorageDealActive
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_FAILING:
		status = storagemarket.StorageDealFailing
	case rpc.StorageDealStatus_STORAGE_DEAL_STATUS_NOT_FOUND:
		status = storagemarket.StorageDealNotFound
	default:
		status = storagemarket.StorageDealUnknown
	}
	return status, res.Slashed, nil
}

// Watch returns a channel with state changes of indicated proposals.
func (d *Deals) Watch(ctx context.Context, proposals []cid.Cid) (<-chan WatchEvent, error) {
	channel := make(chan WatchEvent)
	proposalStrings := make([]string, len(proposals))
	for i, proposal := range proposals {
		proposalStrings[i] = proposal.String()
	}
	stream, err := d.client.Watch(ctx, &rpc.WatchRequest{Proposals: proposalStrings})
	if err != nil {
		return nil, fmt.Errorf("calling Watch: %v", err)
	}
	go func() {
		defer close(channel)
		for {
			event, err := stream.Recv()
			if err != nil {
				stat := status.Convert(err)
				if stat == nil || (stat.Code() != codes.Canceled) {
					channel <- WatchEvent{Err: fmt.Errorf("reveiving stream: %v", err)}
				}
				break
			}
			proposalCid, err := cid.Decode(event.GetDealInfo().GetProposalCid())
			if err != nil {
				channel <- WatchEvent{Err: fmt.Errorf("decoding cid: %v", err)}
				break
			}
			cid, err := cid.Decode(event.GetDealInfo().GetPieceCid())
			if err != nil {
				channel <- WatchEvent{Err: fmt.Errorf("decoding cid: %v", err)}
				break
			}
			deal := deals.StorageDealInfo{
				ProposalCid:     proposalCid,
				StateID:         event.GetDealInfo().GetStateId(),
				StateName:       event.GetDealInfo().GetStateName(),
				Miner:           event.GetDealInfo().GetMiner(),
				PieceCID:        cid,
				Size:            event.GetDealInfo().GetSize(),
				PricePerEpoch:   event.GetDealInfo().GetPricePerEpoch(),
				StartEpoch:      event.GetDealInfo().GetStartEpoch(),
				Duration:        event.GetDealInfo().GetDuration(),
				DealID:          event.GetDealInfo().GetDealId(),
				ActivationEpoch: event.GetDealInfo().GetActivationEpoch(),
				Message:         event.GetDealInfo().GetMsg(),
			}
			channel <- WatchEvent{Deal: deal}
		}
	}()
	return channel, nil
}

// ListStorageDealRecords returns a list of storage deals according to the provided options.
func (d *Deals) ListStorageDealRecords(ctx context.Context, opts ...ListDealRecordsOption) ([]deals.StorageDealRecord, error) {
	conf := &rpc.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	res, err := d.client.ListStorageDealRecords(ctx, &rpc.ListStorageDealRecordsRequest{Config: conf})
	if err != nil {
		return nil, fmt.Errorf("calling ListStorageDealRecords: %v", err)
	}
	ret, err := fromRPCStorageDealRecords(res.Records)
	if err != nil {
		return nil, fmt.Errorf("processing response deal records: %v", err)
	}
	return ret, nil
}

// ListRetrievalDealRecords returns a list of retrieval deals according to the provided options.
func (d *Deals) ListRetrievalDealRecords(ctx context.Context, opts ...ListDealRecordsOption) ([]deals.RetrievalDealRecord, error) {
	conf := &rpc.ListDealRecordsConfig{}
	for _, opt := range opts {
		opt(conf)
	}
	res, err := d.client.ListRetrievalDealRecords(ctx, &rpc.ListRetrievalDealRecordsRequest{Config: conf})
	if err != nil {
		return nil, fmt.Errorf("calling ListRetrievalDealRecords: %v", err)
	}
	ret, err := fromRPCRetrievalDealRecords(res.Records)
	if err != nil {
		return nil, fmt.Errorf("processing response deal records: %v", err)
	}
	return ret, nil
}

func fromRPCStorageDealRecords(records []*rpc.StorageDealRecord) ([]deals.StorageDealRecord, error) {
	var ret []deals.StorageDealRecord
	for _, rpcRecord := range records {
		if rpcRecord.DealInfo == nil {
			continue
		}
		rootCid, err := cid.Decode(rpcRecord.RootCid)
		if err != nil {
			return nil, err
		}
		record := deals.StorageDealRecord{
			RootCid: rootCid,
			Addr:    rpcRecord.Addr,
			Time:    rpcRecord.Time,
			Pending: rpcRecord.Pending,
		}
		proposalCid, err := cid.Decode(rpcRecord.DealInfo.ProposalCid)
		if err != nil {
			return nil, err
		}
		pieceCid, err := cid.Decode(rpcRecord.DealInfo.PieceCid)
		if err != nil {
			return nil, err
		}
		record.DealInfo = deals.StorageDealInfo{
			ProposalCid:     proposalCid,
			StateID:         rpcRecord.DealInfo.StateId,
			StateName:       rpcRecord.DealInfo.StateName,
			Miner:           rpcRecord.DealInfo.Miner,
			PieceCID:        pieceCid,
			Size:            rpcRecord.DealInfo.Size,
			PricePerEpoch:   rpcRecord.DealInfo.PricePerEpoch,
			StartEpoch:      rpcRecord.DealInfo.StartEpoch,
			Duration:        rpcRecord.DealInfo.Duration,
			DealID:          rpcRecord.DealInfo.DealId,
			ActivationEpoch: rpcRecord.DealInfo.ActivationEpoch,
			Message:         rpcRecord.DealInfo.Msg,
		}
		ret = append(ret, record)
	}
	return ret, nil
}

func fromRPCRetrievalDealRecords(records []*rpc.RetrievalDealRecord) ([]deals.RetrievalDealRecord, error) {
	var ret []deals.RetrievalDealRecord
	for _, rpcRecord := range records {
		if rpcRecord.DealInfo == nil {
			continue
		}
		record := deals.RetrievalDealRecord{
			Addr: rpcRecord.Addr,
			Time: rpcRecord.Time,
		}
		rootCid, err := cid.Decode(rpcRecord.DealInfo.RootCid)
		if err != nil {
			return nil, err
		}
		record.DealInfo = deals.RetrievalDealInfo{
			RootCid:                 rootCid,
			Size:                    rpcRecord.DealInfo.Size,
			MinPrice:                rpcRecord.DealInfo.MinPrice,
			PaymentInterval:         rpcRecord.DealInfo.PaymentInterval,
			PaymentIntervalIncrease: rpcRecord.DealInfo.PaymentIntervalIncrease,
			Miner:                   rpcRecord.DealInfo.Miner,
			MinerPeerID:             rpcRecord.DealInfo.MinerPeerId,
		}
		ret = append(ret, record)
	}
	return ret, nil
}
