package powergate

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// ListStorageDealRecords calls ffs.ListStorageDealRecords.
func (s *Service) ListStorageDealRecords(ctx context.Context, req *proto.ListStorageDealRecordsRequest) (*proto.ListStorageDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.ListStorageDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &proto.ListStorageDealRecordsResponse{Records: toRPCStorageDealRecords(records)}, nil
}

// ListRetrievalDealRecords calls ffs.ListRetrievalDealRecords.
func (s *Service) ListRetrievalDealRecords(ctx context.Context, req *proto.ListRetrievalDealRecordsRequest) (*proto.ListRetrievalDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.ListRetrievalDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &proto.ListRetrievalDealRecordsResponse{Records: toRPCRetrievalDealRecords(records)}, nil
}
