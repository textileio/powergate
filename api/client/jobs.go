package client

import (
	"context"

	proto "github.com/textileio/powergate/proto/powergate/v1"
)

// Jobs provides access to Powergate jobs APIs.
type Jobs struct {
	client proto.PowergateServiceClient
}

// StorageConfigForJob returns the StorageConfig associated with the specified job.
func (j *Jobs) StorageConfigForJob(ctx context.Context, jobID string) (*proto.StorageConfigForJobResponse, error) {
	return j.client.StorageConfigForJob(ctx, &proto.StorageConfigForJobRequest{JobId: jobID})
}
