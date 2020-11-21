package user

import (
	"fmt"

	"github.com/ipfs/go-cid"
	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/deals"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/util"
)

// ToRPCStorageConfig converts from a ffs.StorageConfig to a rpc StorageConfig.
func ToRPCStorageConfig(config ffs.StorageConfig) *userPb.StorageConfig {
	return &userPb.StorageConfig{
		Repairable: config.Repairable,
		Hot:        toRPCHotConfig(config.Hot),
		Cold:       toRPCColdConfig(config.Cold),
	}
}

func toRPCHotConfig(config ffs.HotConfig) *userPb.HotConfig {
	return &userPb.HotConfig{
		Enabled:          config.Enabled,
		AllowUnfreeze:    config.AllowUnfreeze,
		UnfreezeMaxPrice: config.UnfreezeMaxPrice,
		Ipfs: &userPb.IpfsConfig{
			AddTimeout: int64(config.Ipfs.AddTimeout),
		},
	}
}

func toRPCColdConfig(config ffs.ColdConfig) *userPb.ColdConfig {
	return &userPb.ColdConfig{
		Enabled: config.Enabled,
		Filecoin: &userPb.FilConfig{
			ReplicationFactor: int64(config.Filecoin.RepFactor),
			DealMinDuration:   config.Filecoin.DealMinDuration,
			ExcludedMiners:    config.Filecoin.ExcludedMiners,
			TrustedMiners:     config.Filecoin.TrustedMiners,
			CountryCodes:      config.Filecoin.CountryCodes,
			Renew: &userPb.FilRenew{
				Enabled:   config.Filecoin.Renew.Enabled,
				Threshold: int64(config.Filecoin.Renew.Threshold),
			},
			Address:         config.Filecoin.Addr,
			MaxPrice:        config.Filecoin.MaxPrice,
			FastRetrieval:   config.Filecoin.FastRetrieval,
			DealStartOffset: config.Filecoin.DealStartOffset,
		},
	}
}

func toRPCDealErrors(des []ffs.DealError) []*userPb.DealError {
	ret := make([]*userPb.DealError, len(des))
	for i, de := range des {
		var strProposalCid string
		if de.ProposalCid.Defined() {
			strProposalCid = util.CidToString(de.ProposalCid)
		}
		ret[i] = &userPb.DealError{
			ProposalCid: strProposalCid,
			Miner:       de.Miner,
			Message:     de.Message,
		}
	}
	return ret
}

func fromRPCHotConfig(config *userPb.HotConfig) ffs.HotConfig {
	res := ffs.HotConfig{}
	if config != nil {
		res.Enabled = config.Enabled
		res.AllowUnfreeze = config.AllowUnfreeze
		res.UnfreezeMaxPrice = config.UnfreezeMaxPrice
		if config.Ipfs != nil {
			ipfs := ffs.IpfsConfig{
				AddTimeout: int(config.Ipfs.AddTimeout),
			}
			res.Ipfs = ipfs
		}
	}
	return res
}

func fromRPCColdConfig(config *userPb.ColdConfig) ffs.ColdConfig {
	res := ffs.ColdConfig{}
	if config != nil {
		res.Enabled = config.Enabled
		if config.Filecoin != nil {
			filecoin := ffs.FilConfig{
				RepFactor:       int(config.Filecoin.ReplicationFactor),
				DealMinDuration: config.Filecoin.DealMinDuration,
				ExcludedMiners:  config.Filecoin.ExcludedMiners,
				CountryCodes:    config.Filecoin.CountryCodes,
				TrustedMiners:   config.Filecoin.TrustedMiners,
				Addr:            config.Filecoin.Address,
				MaxPrice:        config.Filecoin.MaxPrice,
				FastRetrieval:   config.Filecoin.FastRetrieval,
				DealStartOffset: config.Filecoin.DealStartOffset,
			}
			if config.Filecoin.Renew != nil {
				renew := ffs.FilRenew{
					Enabled:   config.Filecoin.Renew.Enabled,
					Threshold: int(config.Filecoin.Renew.Threshold),
				}
				filecoin.Renew = renew
			}
			res.Filecoin = filecoin
		}
	}
	return res
}

func toRPCStorageInfo(info ffs.StorageInfo) *userPb.StorageInfo {
	storageInfo := &userPb.StorageInfo{
		JobId:   info.JobID.String(),
		Cid:     util.CidToString(info.Cid),
		Created: info.Created.UnixNano(),
		Hot: &userPb.HotInfo{
			Enabled: info.Hot.Enabled,
			Size:    int64(info.Hot.Size),
			Ipfs: &userPb.IpfsHotInfo{
				Created: info.Hot.Ipfs.Created.UnixNano(),
			},
		},
		Cold: &userPb.ColdInfo{
			Enabled: info.Cold.Enabled,
			Filecoin: &userPb.FilInfo{
				DataCid:   util.CidToString(info.Cold.Filecoin.DataCid),
				Size:      info.Cold.Filecoin.Size,
				Proposals: make([]*userPb.FilStorage, len(info.Cold.Filecoin.Proposals)),
			},
		},
	}
	for i, p := range info.Cold.Filecoin.Proposals {
		var strProposalCid string
		if p.ProposalCid.Defined() {
			strProposalCid = util.CidToString(p.ProposalCid)
		}
		var strPieceCid string
		if p.PieceCid.Defined() {
			strPieceCid = util.CidToString(p.PieceCid)
		}
		storageInfo.Cold.Filecoin.Proposals[i] = &userPb.FilStorage{
			ProposalCid:     strProposalCid,
			PieceCid:        strPieceCid,
			Renewed:         p.Renewed,
			Duration:        p.Duration,
			ActivationEpoch: p.ActivationEpoch,
			StartEpoch:      p.StartEpoch,
			Miner:           p.Miner,
			EpochPrice:      p.EpochPrice,
		}
	}
	return storageInfo
}

func buildListDealRecordsOptions(conf *userPb.DealRecordsConfig) []deals.DealRecordsOption {
	var opts []deals.DealRecordsOption
	if conf != nil {
		opts = []deals.DealRecordsOption{
			deals.WithAscending(conf.Ascending),
			deals.WithDataCids(conf.DataCids...),
			deals.WithFromAddrs(conf.FromAddrs...),
			deals.WithIncludePending(conf.IncludePending),
			deals.WithIncludeFinal(conf.IncludeFinal),
		}
	}
	return opts
}

func toRPCStorageDealRecords(records []deals.StorageDealRecord) []*userPb.StorageDealRecord {
	ret := make([]*userPb.StorageDealRecord, len(records))
	for i, r := range records {
		ret[i] = &userPb.StorageDealRecord{
			RootCid: util.CidToString(r.RootCid),
			Address: r.Addr,
			Time:    r.Time,
			Pending: r.Pending,
			DealInfo: &userPb.StorageDealInfo{
				ProposalCid:     util.CidToString(r.DealInfo.ProposalCid),
				StateId:         r.DealInfo.StateID,
				StateName:       r.DealInfo.StateName,
				Miner:           r.DealInfo.Miner,
				PieceCid:        util.CidToString(r.DealInfo.PieceCID),
				Size:            r.DealInfo.Size,
				PricePerEpoch:   r.DealInfo.PricePerEpoch,
				StartEpoch:      r.DealInfo.StartEpoch,
				Duration:        r.DealInfo.Duration,
				DealId:          r.DealInfo.DealID,
				ActivationEpoch: r.DealInfo.ActivationEpoch,
				Message:         r.DealInfo.Message,
			},
		}
	}
	return ret
}

func toRPCRetrievalDealRecords(records []deals.RetrievalDealRecord) []*userPb.RetrievalDealRecord {
	ret := make([]*userPb.RetrievalDealRecord, len(records))
	for i, r := range records {
		ret[i] = &userPb.RetrievalDealRecord{
			Address: r.Addr,
			Time:    r.Time,
			DealInfo: &userPb.RetrievalDealInfo{
				RootCid:                 util.CidToString(r.DealInfo.RootCid),
				Size:                    r.DealInfo.Size,
				MinPrice:                r.DealInfo.MinPrice,
				PaymentInterval:         r.DealInfo.PaymentInterval,
				PaymentIntervalIncrease: r.DealInfo.PaymentIntervalIncrease,
				Miner:                   r.DealInfo.Miner,
				MinerPeerId:             r.DealInfo.MinerPeerID,
			},
		}
	}
	return ret
}

// ToProtoStorageJobs converts a slice of ffs.StorageJobs to proto Jobs.
func ToProtoStorageJobs(jobs []ffs.StorageJob) ([]*userPb.StorageJob, error) {
	var res []*userPb.StorageJob
	for _, job := range jobs {
		j, err := toRPCJob(job)
		if err != nil {
			return nil, err
		}
		res = append(res, j)
	}
	return res, nil
}

func toRPCJob(job ffs.StorageJob) (*userPb.StorageJob, error) {
	var dealInfo []*userPb.DealInfo
	for _, item := range job.DealInfo {
		info := &userPb.DealInfo{
			ActivationEpoch: item.ActivationEpoch,
			DealId:          item.DealID,
			Duration:        item.Duration,
			Message:         item.Message,
			Miner:           item.Miner,
			PieceCid:        item.PieceCID.String(),
			PricePerEpoch:   item.PricePerEpoch,
			ProposalCid:     item.ProposalCid.String(),
			Size:            item.Size,
			StartEpoch:      item.StartEpoch,
			StateId:         item.StateID,
			StateName:       item.StateName,
		}
		dealInfo = append(dealInfo, info)
	}

	var status userPb.JobStatus
	switch job.Status {
	case ffs.Unspecified:
		status = userPb.JobStatus_JOB_STATUS_UNSPECIFIED
	case ffs.Queued:
		status = userPb.JobStatus_JOB_STATUS_QUEUED
	case ffs.Executing:
		status = userPb.JobStatus_JOB_STATUS_EXECUTING
	case ffs.Failed:
		status = userPb.JobStatus_JOB_STATUS_FAILED
	case ffs.Canceled:
		status = userPb.JobStatus_JOB_STATUS_CANCELED
	case ffs.Success:
		status = userPb.JobStatus_JOB_STATUS_SUCCESS
	default:
		return nil, fmt.Errorf("unknown job status: %v", job.Status)
	}
	return &userPb.StorageJob{
		Id:         job.ID.String(),
		ApiId:      job.APIID.String(),
		Cid:        util.CidToString(job.Cid),
		Status:     status,
		ErrorCause: job.ErrCause,
		DealErrors: toRPCDealErrors(job.DealErrors),
		CreatedAt:  job.CreatedAt,
		DealInfo:   dealInfo,
	}, nil
}

func fromProtoCids(cids []string) ([]cid.Cid, error) {
	var res []cid.Cid
	for _, cid := range cids {
		cid, err := util.CidFromString(cid)
		if err != nil {
			return nil, err
		}
		res = append(res, cid)
	}
	return res, nil
}
