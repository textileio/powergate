package user

import (
	"context"

	"github.com/ipfs/go-cid"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/v2/api/server/util"
	"github.com/textileio/powergate/v2/ffs/api"
	"github.com/textileio/powergate/v2/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageInfo returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns an error with codes.NotFound.
func (s *Service) StorageInfo(ctx context.Context, req *userPb.StorageInfoRequest) (*userPb.StorageInfoResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}
	cid, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
	}
	info, err := i.StorageInfo(cid)
	if err == api.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "getting storage info: %v", err)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting storage info: %v", err)
	}
	pbInfo := su.ToRPCStorageInfo(info)
	return &userPb.StorageInfoResponse{StorageInfo: pbInfo}, nil
}

// ListStorageInfo returns a list of information about all stored cids, filtered by cids if provided.
func (s *Service) ListStorageInfo(ctx context.Context, req *userPb.ListStorageInfoRequest) (*userPb.ListStorageInfoResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "getting instance: %v", err)
	}
	cids := make([]cid.Cid, len(req.Cids))
	for i, s := range req.Cids {
		c, err := util.CidFromString(s)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
		}
		cids[i] = c
	}
	infos, err := i.ListStorageInfo(cids...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "querying storage info: %v", err)
	}
	res := make([]*userPb.StorageInfo, len(infos))
	for i, info := range infos {
		res[i] = su.ToRPCStorageInfo(info)
	}
	return &userPb.ListStorageInfoResponse{StorageInfo: res}, nil
}
