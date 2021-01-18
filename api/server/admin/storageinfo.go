package admin

import (
	"context"

	"github.com/ipfs/go-cid"
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/v2/api/server/util"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/ffs/api"
	"github.com/textileio/powergate/v2/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageInfo returns the information about a stored Cid. If no information is available,
// since the Cid was never stored, it returns an error with codes.NotFound.
func (a *Service) StorageInfo(ctx context.Context, req *adminPb.StorageInfoRequest) (*adminPb.StorageInfoResponse, error) {
	iid := ffs.APIID(req.UserId)
	cid, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
	}
	info, err := a.s.GetStorageInfo(iid, cid)
	if err == api.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "getting storage info: %v", err)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting storage info: %v", err)
	}
	pbInfo := su.ToRPCStorageInfo(info)
	return &adminPb.StorageInfoResponse{StorageInfo: pbInfo}, nil
}

// ListStorageInfo returns a list of information about all stored cids, filtered by user ids and cids if provided.
func (a *Service) ListStorageInfo(ctx context.Context, req *adminPb.ListStorageInfoRequest) (*adminPb.ListStorageInfoResponse, error) {
	iids := make([]ffs.APIID, len(req.UserIds))
	for i, id := range req.UserIds {
		iids[i] = ffs.APIID(id)
	}
	cids := make([]cid.Cid, len(req.Cids))
	for i, s := range req.Cids {
		c, err := util.CidFromString(s)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
		}
		cids[i] = c
	}
	infos, err := a.s.ListStorageInfo(iids, cids)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "querying storage info: %v", err)
	}
	res := make([]*userPb.StorageInfo, len(infos))
	for i, info := range infos {
		res[i] = su.ToRPCStorageInfo(info)
	}
	return &adminPb.ListStorageInfoResponse{StorageInfo: res}, nil
}
