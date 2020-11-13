package user

import (
	"context"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
)

// StorageDealRecords calls ffs.ListStorageDealRecords.
func (s *Service) StorageDealRecords(ctx context.Context, req *userPb.StorageDealRecordsRequest) (*userPb.StorageDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.StorageDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &userPb.StorageDealRecordsResponse{Records: toRPCStorageDealRecords(records)}, nil
}

// RetrievalDealRecords calls ffs.ListRetrievalDealRecords.
func (s *Service) RetrievalDealRecords(ctx context.Context, req *userPb.RetrievalDealRecordsRequest) (*userPb.RetrievalDealRecordsResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}
	records, err := i.RetrievalDealRecords(buildListDealRecordsOptions(req.Config)...)
	if err != nil {
		return nil, err
	}
	return &userPb.RetrievalDealRecordsResponse{Records: toRPCRetrievalDealRecords(records)}, nil
}
