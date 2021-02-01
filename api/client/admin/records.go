package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

type Records struct {
	client adminPb.AdminServiceClient
}

func (c *Records) GetUpdatedRetrievalRecordsSince(ctx context.Context, sinceNano int64, limit int) (*adminPb.GetUpdatedRetrievalRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedRetrievalRecordsSinceRequest{
		SinceNano: int64(sinceNano),
		Limit:     int32(limit),
	}
	return c.client.GetUpdatedRetrievalRecordsSince(ctx, req)
}

func (c *Records) GetUpdatedStorageDealRecordsSince(ctx context.Context, sinceNano int64, limit int) (*adminPb.GetUpdatedStorageDealRecordsSinceResponse, error) {
	req := &adminPb.GetUpdatedStorageDealRecordsSinceRequest{
		SinceNano: int64(sinceNano),
		Limit:     int32(limit),
	}
	return c.client.GetUpdatedStorageDealRecordsSince(ctx, req)
}
