package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"github.com/textileio/powergate/v2/api/server/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUpdatedStorageDealRecordsSince returns all the storage deal records that got created or updated
// since a provided point in time.
func (a *Service) GetUpdatedStorageDealRecordsSince(ctx context.Context, req *adminPb.GetUpdatedStorageDealRecordsSinceRequest) (*adminPb.GetUpdatedStorageDealRecordsSinceResponse, error) {
	rs, err := a.dm.GetUpdatedStorageDealRecordsSince(req.SinceNano, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting updated storage deal records: %s", err)
	}

	return &adminPb.GetUpdatedStorageDealRecordsSinceResponse{Records: util.ToRPCStorageDealRecords(rs)}, nil
}

// GetUpdatedRetrievalRecordsSince returns all the retrieval records that got created or updated
// since a provided point in time.
func (a *Service) GetUpdatedRetrievalRecordsSince(ctx context.Context, req *adminPb.GetUpdatedRetrievalRecordsSinceRequest) (*adminPb.GetUpdatedRetrievalRecordsSinceResponse, error) {
	rs, err := a.dm.GetUpdatedRetrievalRecordsSince(req.SinceNano, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting updated storage deal records: %s", err)
	}

	return &adminPb.GetUpdatedRetrievalRecordsSinceResponse{Records: util.ToRPCRetrievalDealRecords(rs)}, nil
}
