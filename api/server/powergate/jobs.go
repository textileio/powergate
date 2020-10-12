package powergate

import (
	"context"

	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/rpc"
	"github.com/textileio/powergate/ffs/scheduler"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageConfigForJob returns the StorageConfig associated with the job id.
func (p *Service) StorageConfigForJob(ctx context.Context, req *proto.StorageConfigForJobRequest) (*proto.StorageConfigForJobResponse, error) {
	sc, err := p.s.StorageConfig(ffs.JobID(req.JobId))
	if err != nil {
		code := codes.Internal
		if err == scheduler.ErrNotFound {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "creating lotus client: %v", err)
	}
	res := rpc.ToRPCStorageConfig(sc)
	return &proto.StorageConfigForJobResponse{StorageConfig: res}, nil
}
