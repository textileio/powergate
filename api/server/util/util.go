package util

import (
	"fmt"

	"github.com/ipfs/go-cid"
	userPb "github.com/textileio/powergate/v2/api/gen/powergate/user/v1"
	"github.com/textileio/powergate/v2/deals"
	"github.com/textileio/powergate/v2/ffs"
	"github.com/textileio/powergate/v2/util"
)

// ToRPCStorageInfo converts a StorageInfo to the proto version.
func ToRPCStorageInfo(info ffs.StorageInfo) *userPb.StorageInfo {
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
		var strPieceCid string
		if p.PieceCid.Defined() {
			strPieceCid = util.CidToString(p.PieceCid)
		}
		storageInfo.Cold.Filecoin.Proposals[i] = &userPb.FilStorage{
			DealId:     int64(p.DealID),
			PieceCid:   strPieceCid,
			Renewed:    p.Renewed,
			Duration:   p.Duration,
			StartEpoch: p.StartEpoch,
			Miner:      p.Miner,
			EpochPrice: p.EpochPrice,
		}
	}
	return storageInfo
}

// ToProtoStorageJobs converts a slice of ffs.StorageJobs to proto Jobs.
func ToProtoStorageJobs(jobs []ffs.StorageJob) ([]*userPb.StorageJob, error) {
	var res []*userPb.StorageJob
	for _, job := range jobs {
		j, err := ToRPCJob(job)
		if err != nil {
			return nil, err
		}
		res = append(res, j)
	}
	return res, nil
}

// ToRPCJob converts a job to a proto job.
func ToRPCJob(job ffs.StorageJob) (*userPb.StorageJob, error) {
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

// FromProtoCids converts string cids to cid.Cids.
func FromProtoCids(cids []string) ([]cid.Cid, error) {
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

// ToRPCStorageDealRecords transforms a StorageDealRecord slice to the proto version.
func ToRPCStorageDealRecords(records []deals.StorageDealRecord) []*userPb.StorageDealRecord {
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
			TransferSize:      r.TransferSize,
			DataTransferStart: r.DataTransferStart,
			DataTransferEnd:   r.DataTransferEnd,
			SealingStart:      r.SealingStart,
			SealingEnd:        r.SealingEnd,
			ErrMsg:            r.ErrMsg,
			UpdatedAt:         r.UpdatedAt,
		}
	}
	return ret
}

// ToRPCRetrievalDealRecords converts a RetrievalDealRecord slice to the proto version.
func ToRPCRetrievalDealRecords(records []deals.RetrievalDealRecord) []*userPb.RetrievalDealRecord {
	ret := make([]*userPb.RetrievalDealRecord, len(records))
	for i, r := range records {
		ret[i] = &userPb.RetrievalDealRecord{
			Id:      r.ID,
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
			DataTransferStart: r.DataTransferStart,
			DataTransferEnd:   r.DataTransferEnd,
			ErrMsg:            r.ErrMsg,
			UpdatedAt:         r.UpdatedAt,
		}
	}
	return ret
}
