package admin

import (
	"context"

	proto "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// Data provides access to Powergate data admin APIs.
type Data struct {
	client proto.AdminServiceClient
}

// GCStaged unpins staged data not related to queued or executing jobs.
func (w *Data) GCStaged(ctx context.Context) (*proto.GCStagedResponse, error) {
	return w.client.GCStaged(ctx, &proto.GCStagedRequest{})
}

// PinnedCids returns pinned cids information of hot-storage.
func (w *Data) PinnedCids(ctx context.Context) (*proto.PinnedCidsResponse, error) {
	return w.client.PinnedCids(ctx, &proto.PinnedCidsRequest{})
}
