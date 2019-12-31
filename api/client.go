package api

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
	ma "github.com/multiformats/go-multiaddr"
	pb "github.com/textileio/filecoin/api/pb"
	"github.com/textileio/filecoin/deals"
	"github.com/textileio/filecoin/lotus/types"
	"github.com/textileio/filecoin/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client provides the client api
type Client struct {
	client pb.APIClient
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
}

// WatchEvent is used to send data or error values for Watch
type WatchEvent struct {
	Deal deals.DealInfo
	Err  error
}

// NewClient starts the client
func NewClient(maddr ma.Multiaddr) (*Client, error) {
	addr, err := util.TCPAddrFromMultiAddr(maddr)
	if err != nil {
		return nil, err
	}
	// ToDo: Support secure connection
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		client: pb.NewAPIClient(conn),
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}
	return client, nil
}

// Close closes the client's grpc connection and cancels any active requests
func (c *Client) Close() error {
	c.cancel()
	return c.conn.Close()
}

// AvailableAsks executes a query to retrieve active Asks
func (c *Client) AvailableAsks(query deals.Query) ([]deals.StorageAsk, error) {
	q := &pb.Query{
		MaxPrice:  query.MaxPrice,
		PieceSize: query.PieceSize,
		Limit:     int32(query.Limit),
		Offset:    int32(query.Offset),
	}
	reply, err := c.client.AvailableAsks(c.ctx, &pb.AvailableAsksRequest{Query: q})
	if err != nil {
		return nil, err
	}
	asks := make([]deals.StorageAsk, len(reply.GetAsks()))
	for i, ask := range reply.GetAsks() {
		asks[i] = deals.StorageAsk{
			Price:        ask.GetPrice(),
			MinPieceSize: ask.GetMinPieceSize(),
			Miner:        ask.GetMiner(),
			Timestamp:    ask.GetTimestamp(),
			Expiry:       ask.GetExpiry(),
		}
	}
	return asks, nil
}

// Store creates a proposal deal for data using wallet addr to all miners indicated
// by dealConfigs for duration epochs
func (c *Client) Store(addr string, data io.Reader, dealConfigs []deals.DealConfig, duration uint64) ([]cid.Cid, []deals.DealConfig, error) {
	stream, err := c.client.Store(c.ctx)
	if err != nil {
		return nil, nil, err
	}

	reqDealConfigs := make([]*pb.DealConfig, len(dealConfigs))
	for i, dealConfig := range dealConfigs {
		reqDealConfigs[i] = &pb.DealConfig{
			Miner:      dealConfig.Miner,
			EpochPrice: dealConfig.EpochPrice.Uint64(),
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

	cids := make([]cid.Cid, len(reply.GetCids()))
	for i, replyCid := range reply.GetCids() {
		id, err := cid.Decode(replyCid)
		if err != nil {
			return nil, nil, err
		}
		cids[i] = id
	}

	failedDeals := make([]deals.DealConfig, len(reply.GetFailedDeals()))
	for i, dealConfig := range reply.GetFailedDeals() {
		failedDeals[i] = deals.DealConfig{
			Miner:      dealConfig.GetMiner(),
			EpochPrice: types.NewInt(dealConfig.GetEpochPrice()),
		}
	}

	return cids, failedDeals, nil
}

// Watch returnas a channel with state changes of indicated proposals
func (c *Client) Watch(proposals []cid.Cid) (<-chan WatchEvent, context.CancelFunc, error) {
	channel := make(chan WatchEvent)
	ctx, cancel := context.WithCancel(c.ctx)
	proposalStrings := make([]string, len(proposals))
	for i, proposal := range proposals {
		proposalStrings[i] = proposal.String()
	}
	stream, err := c.client.Watch(ctx, &pb.WatchRequest{Proposals: proposalStrings})
	if err != nil {
		cancel()
		return nil, nil, err
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
				PricePerEpoch: types.NewInt(event.GetDealInfo().GetPricePerEpoch()),
				Duration:      event.GetDealInfo().GetDuration(),
			}
			channel <- WatchEvent{Deal: deal}
		}
	}()
	return channel, cancel, nil
}
