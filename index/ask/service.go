package ask

import (
	"context"

	pb "github.com/textileio/fil-tools/index/ask/pb"
)

// Service implements the gprc service
type Service struct {
	pb.UnimplementedAPIServer

	index *AskIndex
}

// NewService is a helper to create a new Service
func NewService(ai *AskIndex) *Service {
	return &Service{
		index: ai,
	}
}

// Get calls askIndex.Get
func (s *Service) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetReply, error) {
	index := s.index.Get()
	storage := make(map[string]*pb.StorageAsk, len(index.Storage))
	for key, ask := range index.Storage {
		storage[key] = &pb.StorageAsk{
			Price:        ask.Price,
			MinPieceSize: ask.MinPieceSize,
			Miner:        ask.Miner,
			Timestamp:    ask.Timestamp,
			Expiry:       ask.Expiry,
		}
	}
	pbIndex := &pb.Index{
		LastUpdated:        index.LastUpdated.Unix(),
		StorageMedianPrice: index.StorageMedianPrice,
		Storage:            storage,
	}
	return &pb.GetReply{Index: pbIndex}, nil
}

// Query calls askIndex.Query
func (s *Service) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryReply, error) {
	q := Query{
		MaxPrice:  req.GetQuery().GetMaxPrice(),
		PieceSize: req.GetQuery().GetPieceSize(),
		Limit:     int(req.GetQuery().GetLimit()),
		Offset:    int(req.GetQuery().GetOffset()),
	}
	asks, err := s.index.Query(q)
	if err != nil {
		return nil, err
	}
	replyAsks := make([]*pb.StorageAsk, len(asks))
	for i, ask := range asks {
		replyAsks[i] = &pb.StorageAsk{
			Price:        ask.Price,
			MinPieceSize: ask.MinPieceSize,
			Miner:        ask.Miner,
			Timestamp:    ask.Timestamp,
			Expiry:       ask.Expiry,
		}
	}
	return &pb.QueryReply{Asks: replyAsks}, nil
}
