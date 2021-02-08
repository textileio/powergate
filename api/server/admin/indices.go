package admin

import (
	"context"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

func (s *Service) GetMiners(ctx context.Context, req *adminPb.GetMinersRequest) (*adminPb.GetMinersResponse, error) {
	info := s.mi.Get()

	var resMiners []*adminPb.FilecoinMiner
	for addr, info := range info.OnChain.Miners {
		if req.WithPower && info.Power == 0 {
			continue
		}
		resMiners = append(resMiners, &adminPb.FilecoinMiner{
			Address: addr,
		})
	}

	res := &adminPb.GetMinersResponse{
		Miners: resMiners,
	}

	return res, nil
}

func (s *Service) GetMinerInfo(ctx context.Context, req *adminPb.GetMinerInfoRequest) (*adminPb.GetMinerInfoResponse, error) {
	res := &adminPb.GetMinerInfoResponse{}

	mi := s.mi.Get()
	ai := s.ai.Get()
	for _, minerAddr := range req.Miners {
		onchain, ok := mi.OnChain.Miners[minerAddr]
		if !ok {
			continue
		}

		lastAsk := ai.Storage[minerAddr]
		res.MinersInfo = append(res.MinersInfo, &adminPb.MinerInfo{
			Address:          minerAddr,
			AskPrice:         lastAsk.Price,
			AskVerifiedPrice: lastAsk.VerifiedPrice,
			MaxPieceSize:     lastAsk.MaxPieceSize,
			MinPieceSize:     lastAsk.MinPieceSize,
			RelativePower:    onchain.RelativePower,
			SectorSize:       onchain.SectorSize,
		})
	}

	return res, nil
}
