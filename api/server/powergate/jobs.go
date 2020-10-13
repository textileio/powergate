package powergate

import (
	"context"

	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/ffs/manager"
	"github.com/textileio/powergate/ffs/rpc"
	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageConfigForJob returns the StorageConfig associated with the job id.
func (s *Service) StorageConfigForJob(ctx context.Context, req *proto.StorageConfigForJobRequest) (*proto.StorageConfigForJobResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		code := codes.Internal
		if err == manager.ErrAuthTokenNotFound || err == ErrEmptyAuthToken {
			code = codes.PermissionDenied
		}
		return nil, status.Errorf(code, "getting instance: %v", err)
	}
	sc, err := i.StorageConfigForJob(ffs.JobID(req.JobId))
	if err != nil {
		code := codes.Internal
		if err == api.ErrNotFound {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "getting storage config for job: %v", err)
	}
	res := rpc.ToRPCStorageConfig(sc)
	return &proto.StorageConfigForJobResponse{StorageConfig: res}, nil
}
