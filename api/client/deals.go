package client

import (
	"context"
	"fmt"
	"io"

	"github.com/filecoin-project/go-address"
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

// WithFromAddrs limits the results deals initated from the provided wallet addresses.
func WithFromAddrs(addrs ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.FromAddrs = addrs
	}
}

// WithDataCids limits the results to deals for the provided data cids.
func WithDataCids(cids ...string) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.DataCids = cids
	}
}

// WithIncludePending specifies whether or not to include pending deals in the results.
// Ignored for ListRetrievalDealRecords.
func WithIncludePending(includePending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.IncludePending = includePending
	}
}

// WithOnlyPending specifies whether or not to only include pending deals in the results.
func WithOnlyPending(onlyPending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.OnlyPending = onlyPending
	}
}

// WithAscending specifies to sort the results in ascending order.
// Default is descending order.
// If pending, records are sorted by timestamp, otherwise records
// are sorted by activation epoch then timestamp.
func WithAscending(ascending bool) ListDealRecordsOption {
	return func(c *rpc.ListDealRecordsConfig) {
		c.Ascending = ascending
	}
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs.
func (d *Deals) Store(ctx context.Context, addr string, data io.Reader, dealConfigs []deals.StorageDealConfig, minDuration uint64) ([]cid.Cid, []deals.StorageDealConfig, error) {
	stream, err := d.client.Store(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("calling Store: %v", err)
	}

	reqDealConfigs := make([]*rpc.DealConfig, len(dealConfigs))
	for i, dealConfig := range dealConfigs {
		reqDealConfigs[i] = &rpc.DealConfig{
			Miner:      dealConfig.Miner,
			EpochPrice: dealConfig.EpochPrice,
		}
	}
	storeParams := &rpc.StoreParams{
		Address:     addr,
		DealConfigs: reqDealConfigs,
		MinDuration: minDuration,
	}
	innerReq := &rpc.StoreRequest_StoreParams{StoreParams: storeParams}

	if err = stream.Send(&rpc.StoreRequest{Payload: innerReq}); err != nil {
		return nil, nil, fmt.Errorf("calling Send: %v", err)
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, nil, fmt.Errorf("reading buffer: %v", err)
		}
		sendErr := stream.Send(&rpc.StoreRequest{Payload: &rpc.StoreRequest_Chunk{Chunk: buffer[:bytesRead]}})
		if sendErr != nil {
			return nil, nil, fmt.Errorf("calling Send: %v", err)
		}
		if err == io.EOF {
			break
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, nil, fmt.Errorf("calling CloseAndRecv: %v", err)
	}

	cids := make([]cid.Cid, len(reply.GetProposalCids()))
	for i, replyCid := range reply.GetProposalCids() {
		id, err := cid.Decode(replyCid)
		if err != nil {
			return nil, nil, fmt.Errorf("decoding cid: %v", err)
		}
		cids[i] = id
	}

	failedDeals := make([]deals.StorageDealConfig, len(reply.GetFailedDeals()))
	for i, dealConfig := range reply.GetFailedDeals() {
		addr, err := address.NewFromString(dealConfig.GetMiner())
		if err != nil {
			return nil, nil, fmt.Errorf("decoding address: %v", err)
		}
		failedDeals[i] = deals.StorageDealConfig{
			Miner:      addr.String(),
			EpochPrice: dealConfig.GetEpochPrice(),
		}
	}

	return cids, failedDeals, nil
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

// Retrieve is used to fetch data from filecoin.
func (d *Deals) Retrieve(ctx context.Context, waddr string, cid cid.Cid) (io.Reader, error) {
	req := &rpc.RetrieveRequest{
		Address: waddr,
		Cid:     cid.String(),
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

// ListStorageDealRecords returns a list of storage deals
// according to the provided options.
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

// ListRetrievalDealRecords returns a list of retrieval deals
// according to the provided options.
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
		record := deals.StorageDealRecord{
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
		pieceCid, err := cid.Decode(rpcRecord.DealInfo.PieceCid)
		if err != nil {
			return nil, err
		}
		record.DealInfo = deals.RetrievalDealInfo{
			PieceCID:                pieceCid,
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
