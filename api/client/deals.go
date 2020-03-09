package client

import (
	"context"
	"io"

	"github.com/filecoin-project/go-address"
	cid "github.com/ipfs/go-cid"
	"github.com/textileio/powergate/deals"
	pb "github.com/textileio/powergate/deals/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deals provides an API for managing deals and storing data
type Deals struct {
	client pb.APIClient
}

// WatchEvent is used to send data or error values for Watch
type WatchEvent struct {
	Deal deals.DealInfo
	Err  error
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs
func (d *Deals) Store(ctx context.Context, addr string, data io.Reader, dealConfigs []deals.StorageDealConfig, duration uint64) ([]cid.Cid, []deals.StorageDealConfig, error) {
	stream, err := d.client.Store(ctx)
	if err != nil {
		return nil, nil, err
	}

	reqDealConfigs := make([]*pb.DealConfig, len(dealConfigs))
	for i, dealConfig := range dealConfigs {
		reqDealConfigs[i] = &pb.DealConfig{
			Miner:      dealConfig.Miner,
			EpochPrice: dealConfig.EpochPrice,
		}
	}
	storeParams := &pb.StoreParams{
		Address:     addr,
		DealConfigs: reqDealConfigs,
		Duration:    duration,
	}
	innerReq := &pb.StoreRequest_StoreParams{StoreParams: storeParams}

	if err = stream.Send(&pb.StoreRequest{Payload: innerReq}); err != nil {
		return nil, nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}
		sendErr := stream.Send(&pb.StoreRequest{Payload: &pb.StoreRequest_Chunk{Chunk: buffer[:bytesRead]}})
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

// Watch returnas a channel with state changes of indicated proposals
func (d *Deals) Watch(ctx context.Context, proposals []cid.Cid) (<-chan WatchEvent, error) {
	channel := make(chan WatchEvent)
	proposalStrings := make([]string, len(proposals))
	for i, proposal := range proposals {
		proposalStrings[i] = proposal.String()
	}
	stream, err := d.client.Watch(ctx, &pb.WatchRequest{Proposals: proposalStrings})
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
			deal := deals.DealInfo{
				ProposalCid:   proposalCid,
				StateID:       event.GetDealInfo().GetStateID(),
				StateName:     event.GetDealInfo().GetStateName(),
				Miner:         event.GetDealInfo().GetMiner(),
				PieceRef:      event.GetDealInfo().GetPieceRef(),
				Size:          event.GetDealInfo().GetSize(),
				PricePerEpoch: event.GetDealInfo().GetPricePerEpoch(),
				Duration:      event.GetDealInfo().GetDuration(),
			}
			channel <- WatchEvent{Deal: deal}
		}
	}()
	return channel, nil
}

// Retrieve is used to fetch data from filecoin
func (d *Deals) Retrieve(ctx context.Context, waddr string, cid cid.Cid) (io.Reader, error) {
	req := &pb.RetrieveRequest{
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
