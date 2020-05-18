package rpc

import (
	"context"

	"github.com/textileio/powergate/index/ask"
)

// Service implements the gprc service
type Service struct {
	UnimplementedAPIServer

	index *ask.Index
}

// NewService is a helper to create a new Service
func NewService(ai *ask.Index) *Service {
	return &Service{
		index: ai,
	}
}

// Get calls askIndex.Get
func (s *Service) Get(ctx context.Context, req *GetRequest) (*GetReply, error) {
	index := s.index.Get()
	storage := make(map[string]*StorageAsk, len(index.Storage))
	for key, ask := range index.Storage {
		storage[key] = &StorageAsk{
			Price:        ask.Price,
			MinPieceSize: ask.MinPieceSize,
			Miner:        ask.Miner,
			Timestamp:    ask.Timestamp,
			Expiry:       ask.Expiry,
		}
	}
	pbIndex := &Index{
		LastUpdated:        index.LastUpdated.Unix(),
		StorageMedianPrice: index.StorageMedianPrice,
		Storage:            storage,
	}
	return &GetReply{Index: pbIndex}, nil
}

// Query calls askIndex.Query
func (s *Service) Query(ctx context.Context, req *QueryRequest) (*QueryReply, error) {
	q := ask.Query{
		MaxPrice:  req.GetQuery().GetMaxPrice(),
		PieceSize: req.GetQuery().GetPieceSize(),
		Limit:     int(req.GetQuery().GetLimit()),
		Offset:    int(req.GetQuery().GetOffset()),
	}
	asks, err := s.index.Query(q)
	if err != nil {
		return nil, err
	}
	replyAsks := make([]*StorageAsk, len(asks))
	for i, ask := range asks {
		replyAsks[i] = &StorageAsk{
			Price:        ask.Price,
			MinPieceSize: ask.MinPieceSize,
			Miner:        ask.Miner,
			Timestamp:    ask.Timestamp,
			Expiry:       ask.Expiry,
		}
	}
	return &QueryReply{Asks: replyAsks}, nil
}
