package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Records provides APIs to fetch generated records from the deals module.
type Records struct {
	client adminPb.AdminServiceClient
}

// GetUpdatedRetrievalRecordsSince returns the retrieval records that
// were created or modified since the specified date.
func (c *Records) GetUpdatedRetrievalRecordsSince(ctx context.Context, sinceNano int64, limit int) (*adminPb.GetUpdatedRetrievalRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedRetrievalRecordsSinceRequest{
		SinceNano: sinceNano,
		Limit:     int32(limit),
	}
	return c.client.GetUpdatedRetrievalRecordsSince(ctx, req)
}

// GetUpdatedStorageDealRecordsSince returns the storage-deal records that
// were created or modified since the specified date.
func (c *Records) GetUpdatedStorageDealRecordsSince(ctx context.Context, sinceNano int64, limit int) (*adminPb.GetUpdatedStorageDealRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedStorageDealRecordsSinceRequest{
		SinceNano: sinceNano,
		Limit:     int32(limit),
	}
	return c.client.GetUpdatedStorageDealRecordsSince(ctx, req)
}
