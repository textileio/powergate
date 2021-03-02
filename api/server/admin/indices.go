package admin

import (
	"context"
	"strconv"

	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
)

// GetMiners returns all miner addresses that satisfy the provided filters.
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

// GetMinerInfo return indices information for the provider miners.
func (s *Service) GetMinerInfo(ctx context.Context, req *adminPb.GetMinerInfoRequest) (*adminPb.GetMinerInfoResponse, error) {
	res := &adminPb.GetMinerInfoResponse{}

	mi := s.mi.Get()
	ai := s.ai.Get()
	for _, minerAddr := range req.Miners {
		onchain, ok := mi.OnChain.Miners[minerAddr]
		if !ok {
			continue
		}

		var minerLocation string
		meta, ok := mi.Meta.Info[minerAddr]
		if ok {
			minerLocation = meta.Location.Country
		}

		lastAsk := ai.Storage[minerAddr]
		res.MinersInfo = append(res.MinersInfo, &adminPb.MinerInfo{
			Address:          minerAddr,
			AskPrice:         strconv.FormatUint(lastAsk.Price, 10),
			AskVerifiedPrice: strconv.FormatUint(lastAsk.VerifiedPrice, 10),
			MaxPieceSize:     lastAsk.MaxPieceSize,
			MinPieceSize:     lastAsk.MinPieceSize,
			RelativePower:    onchain.RelativePower,
			SectorSize:       onchain.SectorSize,
			SectorsLive:      onchain.SectorsLive,
			SectorsFaulty:    onchain.SectorsFaulty,
			SectorsActive:    onchain.SectorsActive,
			Location:         minerLocation,
		})
	}

	return res, nil
}
