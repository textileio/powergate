package admin

import (
	"context"

	adminProto "github.com/textileio/powergate/api/gen/powergate/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GCStaged runs a unpinning garbage collection and returns the unpinned cids.
func (a *Service) GCStaged(ctx context.Context, req *adminProto.GCStagedRequest) (*adminProto.GCStagedResponse, error) {
	cids, err := a.s.GCStaged(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "running FFS GC: %v", err)
	}

	cidsStr := make([]string, len(cids))
	for i := range cids {
		cidsStr[i] = cids[i].String()
	}

	return &adminProto.GCStagedResponse{
		UnpinnedCids: cidsStr,
	}, nil
}

// PinnedCids returns all the pinned cids in Hot-Storage.
func (a *Service) PinnedCids(ctx context.Context, req *adminProto.PinnedCidsRequest) (*adminProto.PinnedCidsResponse, error) {
	pcids, err := a.s.PinnedCids(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting pinned cids: %v", err)
	}

	res := &adminProto.PinnedCidsResponse{
		Cids: make([]*adminProto.HSPinnedCid, len(pcids)),
	}

	for i, pc := range pcids {
		hspc := &adminProto.HSPinnedCid{
			Cid:   pc.Cid.String(),
			Users: make([]*adminProto.HSPinnedCidUser, len(pc.APIIDs)),
		}

		for j, up := range pc.APIIDs {
			upr := &adminProto.HSPinnedCidUser{
				UserId:    up.ID.String(),
				Staged:    up.Staged,
				CreatedAt: up.CreatedAt,
			}
			hspc.Users[j] = upr
		}
		res.Cids[i] = hspc
	}

	return res, nil
}
