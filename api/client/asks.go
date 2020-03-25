package client

import (
	"context"
	"time"

	"github.com/textileio/powergate/index/ask"
	pb "github.com/textileio/powergate/index/ask/pb"
)

// Asks provides an API for viewing asks data
type Asks struct {
	client pb.APIClient
}

// Get returns the current index of available asks
func (a *Asks) Get(ctx context.Context) (*ask.Index, error) {
	reply, err := a.client.Get(ctx, &pb.GetRequest{})
	if err != nil {
		return nil, err
	}
	lastUpdated := time.Unix(reply.GetIndex().GetLastUpdated(), 0)
	storage := make(map[string]ask.StorageAsk, len(reply.GetIndex().GetStorage()))
	for key, val := range reply.GetIndex().GetStorage() {
		storage[key] = askFromPbAsk(val)
	}
	return &ask.Index{
		LastUpdated:        lastUpdated,
		StorageMedianPrice: reply.GetIndex().StorageMedianPrice,
		Storage:            storage,
	}, nil
}

// Query executes a query to retrieve active Asks
func (a *Asks) Query(ctx context.Context, query ask.Query) ([]ask.StorageAsk, error) {
	q := &pb.Query{
		MaxPrice:  query.MaxPrice,
		PieceSize: query.PieceSize,
		Limit:     int32(query.Limit),
		Offset:    int32(query.Offset),
	}
	reply, err := a.client.Query(ctx, &pb.QueryRequest{Query: q})
	if err != nil {
		return nil, err
	}
	asks := make([]ask.StorageAsk, len(reply.GetAsks()))
	for i, a := range reply.GetAsks() {
		asks[i] = askFromPbAsk(a)
	}
	return asks, nil
}

func askFromPbAsk(a *pb.StorageAsk) ask.StorageAsk {
	return ask.StorageAsk{
		Price:        a.GetPrice(),
		MinPieceSize: a.GetMinPieceSize(),
		Miner:        a.GetMiner(),
		Timestamp:    a.GetTimestamp(),
		Expiry:       a.GetExpiry(),
	}
}
