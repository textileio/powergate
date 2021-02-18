package admin

import (
	"context"
	"time"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Records provides APIs to fetch generated records from the deals module.
type Records struct {
	client adminPb.AdminServiceClient
}

// GetUpdatedRetrievalRecordsSince returns the retrieval records that
// were created or modified since the specified date.
func (c *Records) GetUpdatedRetrievalRecordsSince(ctx context.Context, since time.Time, limit int) (*adminPb.GetUpdatedRetrievalRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedRetrievalRecordsSinceRequest{
		Since: timestamppb.New(since),
		Limit: int32(limit),
	}
	return c.client.GetUpdatedRetrievalRecordsSince(ctx, req)
}

// GetUpdatedStorageDealRecordsSince returns the storage-deal records that
// were created or modified since the specified date.
func (c *Records) GetUpdatedStorageDealRecordsSince(ctx context.Context, since time.Time, limit int) (*adminPb.GetUpdatedStorageDealRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedStorageDealRecordsSinceRequest{
		Since: timestamppb.New(since),
		Limit: int32(limit),
	}
	return c.client.GetUpdatedStorageDealRecordsSince(ctx, req)
}
