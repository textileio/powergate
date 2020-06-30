package client

import (
	"context"
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
	Deal deals.DealInfo
	Err  error
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs.
func (d *Deals) Store(ctx context.Context, addr string, data io.Reader, dealConfigs []deals.StorageDealConfig, minDuration uint64) ([]cid.Cid, []deals.StorageDealConfig, error) {
	stream, err := d.client.Store(ctx)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		sendErr := stream.Send(&rpc.StoreRequest{Payload: &rpc.StoreRequest_Chunk{Chunk: buffer[:bytesRead]}})
		if sendErr != nil {
			return nil, nil, sendErr
		}
		if err == io.EOF {
			break
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, nil, err
	}

	cids := make([]cid.Cid, len(reply.GetProposalCids()))
	for i, replyCid := range reply.GetProposalCids() {
		id, err := cid.Decode(replyCid)
		if err != nil {
			return nil, nil, err
		}
		cids[i] = id
	}

	failedDeals := make([]deals.StorageDealConfig, len(reply.GetFailedDeals()))
	for i, dealConfig := range reply.GetFailedDeals() {
		addr, err := address.NewFromString(dealConfig.GetMiner())
		if err != nil {
			return nil, nil, err
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
		return nil, err
	}
	go func() {
		defer close(channel)
		for {
			event, err := stream.Recv()
			if err != nil {
				stat := status.Convert(err)
				if stat == nil || (stat.Code() != codes.Canceled) {
					channel <- WatchEvent{Err: err}
				}
				break
			}
			proposalCid, err := cid.Decode(event.GetDealInfo().GetProposalCid())
			if err != nil {
				channel <- WatchEvent{Err: err}
				break
			}
			cid, err := cid.Decode(event.GetDealInfo().GetPieceCid())
			if err != nil {
				channel <- WatchEvent{Err: err}
				break
			}
			deal := deals.DealInfo{
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
		return nil, err
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

// FinalDealRecords returns a list of all finalized storage deals.
func (d *Deals) FinalDealRecords(ctx context.Context) ([]deals.DealRecord, error) {
	res, err := d.client.FinalDealRecords(ctx, &rpc.FinalDealRecordsRequest{})
	if err != nil {
		return nil, err
	}
	ret, err := fromRPCDealRecords(res.Records)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// PendingDealRecords returns a list of all pending storage deals.
func (d *Deals) PendingDealRecords(ctx context.Context) ([]deals.DealRecord, error) {
	res, err := d.client.PendingDealRecords(ctx, &rpc.PendingDealRecordsRequest{})
	if err != nil {
		return nil, err
	}
	ret, err := fromRPCDealRecords(res.Records)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// AllDealRecords returns a list of all finalized and pending deals.
func (d *Deals) AllDealRecords(ctx context.Context) ([]deals.DealRecord, error) {
	res, err := d.client.AllDealRecords(ctx, &rpc.AllDealRecordsRequest{})
	if err != nil {
		return nil, err
	}
	ret, err := fromRPCDealRecords(res.Records)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// RetrievalRecords returns a list of all retrievals.
func (d *Deals) RetrievalRecords(ctx context.Context) ([]deals.RetrievalRecord, error) {
	res, err := d.client.RetrievalRecords(ctx, &rpc.RetrievalRecordsRequest{})
	if err != nil {
		return nil, err
	}
	ret := make([]deals.RetrievalRecord, len(res.Records))
	for i, rpcRecord := range res.Records {
		record := deals.RetrievalRecord{
			From: rpcRecord.From,
			Time: rpcRecord.Time,
		}
		if rpcRecord.RetrievalInfo != nil {
			pieceCid, err := cid.Decode(rpcRecord.RetrievalInfo.PieceCid)
			if err != nil {
				return nil, err
			}
			record.RetrievalInfo = deals.RetrievalInfo{
				PieceCID:                pieceCid,
				Size:                    rpcRecord.RetrievalInfo.Size,
				MinPrice:                rpcRecord.RetrievalInfo.MinPrice,
				PaymentInterval:         rpcRecord.RetrievalInfo.PaymentInterval,
				PaymentIntervalIncrease: rpcRecord.RetrievalInfo.PaymentIntervalIncrease,
				Miner:                   rpcRecord.RetrievalInfo.Miner,
				MinerPeerID:             rpcRecord.RetrievalInfo.MinerPeerId,
			}
		}
		ret[i] = record
	}
	return ret, nil
}

func fromRPCDealRecords(records []*rpc.DealRecord) ([]deals.DealRecord, error) {
	ret := make([]deals.DealRecord, len(records))
	for i, rpcRecord := range records {
		record := deals.DealRecord{
			From:    rpcRecord.From,
			Time:    rpcRecord.Time,
			Pending: rpcRecord.Pending,
		}
		if rpcRecord.DealInfo != nil {
			proposalCid, err := cid.Decode(rpcRecord.DealInfo.ProposalCid)
			if err != nil {
				return nil, err
			}
			pieceCid, err := cid.Decode(rpcRecord.DealInfo.PieceCid)
			if err != nil {
				return nil, err
			}
			record.DealInfo = deals.DealInfo{
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
		}
		ret[i] = record
	}
	return ret, nil
}
