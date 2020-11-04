package powergate

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// StorageDealRecords calls ffs.ListStorageDealRecords.
func (s *Service) StorageDealRecords(ctx context.Context, req *proto.StorageDealRecordsRequest) (*proto.StorageDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.StorageDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &proto.StorageDealRecordsResponse{Records: toRPCStorageDealRecords(records)}, nil
}

// RetrievalDealRecords calls ffs.ListRetrievalDealRecords.
func (s *Service) RetrievalDealRecords(ctx context.Context, req *proto.RetrievalDealRecordsRequest) (*proto.RetrievalDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.RetrievalDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &proto.RetrievalDealRecordsResponse{Records: toRPCRetrievalDealRecords(records)}, nil
}
