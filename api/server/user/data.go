package user

import (
	"context"
	"fmt"
	"io"
	"sort"

	userPb "github.com/textileio/powergate/api/gen/powergate/user/v1"
	su "github.com/textileio/powergate/api/server/util"
	"github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/api"
	"github.com/textileio/powergate/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Stage allows to stage-pin a stream of data in Hot-Storage in preparation for pushing a storage configuration.
func (s *Service) Stage(srv userPb.UserService_StageServer) error {
	// check that an API instance exists so not just anyone can add data to the hot layer
	fapi, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return fmt.Errorf("getting user instance: %s", err)
	}

	reader, writer := io.Pipe()
	defer func() {
		if err := reader.Close(); err != nil {
			log.Errorf("closing reader: %s", err)
		}
	}()

	go receiveFile(srv, writer)

	c, err := s.hot.Stage(srv.Context(), fapi.ID(), reader)
	if err != nil {
		return fmt.Errorf("adding data to hot storage: %s", err)
	}

	return srv.SendAndClose(&userPb.StageResponse{Cid: util.CidToString(c)})
}

// StageCid allows to stage-pin a cid in Hot-Storage in preparation for pushing a storage configuration.
func (s *Service) StageCid(ctx context.Context, req *userPb.StageCidRequest) (*userPb.StageCidResponse, error) {
	// check that an API instance exists so not just anyone can add data to the hot layer
	fapi, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user instance: %s", err)
	}

	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return nil, fmt.Errorf("parsing cid: %s", err)
	}

	err = s.hot.StageCid(ctx, fapi.ID(), c)
	if err != nil {
		return nil, fmt.Errorf("stage pinning cid in hot-storage: %s", err)
	}

	return &userPb.StageCidResponse{}, nil
}

// ReplaceData calls ffs.Replace.
func (s *Service) ReplaceData(ctx context.Context, req *userPb.ReplaceDataRequest) (*userPb.ReplaceDataResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	c1, err := util.CidFromString(req.Cid1)
	if err != nil {
		return nil, err
	}
	c2, err := util.CidFromString(req.Cid2)
	if err != nil {
		return nil, err
	}

	jid, err := i.Replace(c1, c2)
	if err != nil {
		return nil, err
	}

	return &userPb.ReplaceDataResponse{JobId: jid.String()}, nil
}

// Get gets the data for a stored Cid.
func (s *Service) Get(req *userPb.GetRequest, srv userPb.UserService_GetServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}
	c, err := util.CidFromString(req.GetCid())
	if err != nil {
		return err
	}
	r, err := i.Get(srv.Context(), c)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024*32)
	for {
		bytesRead, err := r.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if sendErr := srv.Send(&userPb.GetResponse{Chunk: buffer[:bytesRead]}); sendErr != nil {
			return sendErr
		}
		if err == io.EOF {
			return nil
		}
	}
}

// WatchLogs returns a stream of human-readable messages related to executions of a Cid.
// The listener is automatically unsubscribed when the client closes the stream.
func (s *Service) WatchLogs(req *userPb.WatchLogsRequest, srv userPb.UserService_WatchLogsServer) error {
	i, err := s.getInstanceByToken(srv.Context())
	if err != nil {
		return err
	}

	opts := []api.GetLogsOption{api.WithHistory(req.History)}
	if req.JobId != ffs.EmptyJobID.String() {
		opts = append(opts, api.WithJidFilter(ffs.JobID(req.JobId)))
	}

	c, err := util.CidFromString(req.Cid)
	if err != nil {
		return err
	}
	ch := make(chan ffs.LogEntry, 100)
	go func() {
		err = i.WatchLogs(srv.Context(), ch, c, opts...)
		close(ch)
	}()
	for l := range ch {
		reply := &userPb.WatchLogsResponse{
			LogEntry: &userPb.LogEntry{
				Cid:     util.CidToString(c),
				JobId:   l.Jid.String(),
				Time:    l.Timestamp.Unix(),
				Message: l.Msg,
			},
		}
		if err := srv.Send(reply); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// CidSummary gives a summary of the storage and jobs state of the specified cid.
func (s *Service) CidSummary(ctx context.Context, req *userPb.CidSummaryRequest) (*userPb.CidSummaryResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	cids, err := fromProtoCids(req.Cids)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cids: %v", err)
	}

	storageConfigs, err := i.GetStorageConfigs(cids...)
	if err != nil {
		code := codes.Internal
		if err == api.ErrNotFound {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "getting storage configs: %v", err)
	}

	type source struct {
		cid            string
		currentStorage *ffs.StorageInfo
		queuedJobs     []ffs.StorageJob
		executingJob   *ffs.StorageJob
	}

	var sources []source

	for cid := range storageConfigs {
		d := source{cid: cid.String()}

		info, err := i.StorageInfo(cid)
		if err != nil && err != api.ErrNotFound {
			return nil, status.Errorf(codes.Internal, "getting storage info: %v", err)
		} else if err == nil {
			d.currentStorage = &info
		}

		d.queuedJobs = i.QueuedStorageJobs(cid)

		executingJobs := i.ExecutingStorageJobs(cid)
		if len(executingJobs) == 1 {
			// There is exactly one job in the slice because we specified a cid
			// and there can be only one executing job per cid at a time.
			d.executingJob = &executingJobs[0]
		} else if len(executingJobs) > 1 {
			log.Warnf("received %d executing jobs when there should be 1", len(executingJobs))
		}

		sources = append(sources, d)
	}

	extractJobTime := func(d source) int64 {
		var time int64
		if d.executingJob != nil {
			time = d.executingJob.CreatedAt
		}
		for _, job := range d.queuedJobs {
			if job.CreatedAt > time {
				time = job.CreatedAt
			}
		}
		return time
	}

	extractCurrentStorageTime := func(d source) int64 {
		var time int64
		if d.currentStorage != nil {
			return d.currentStorage.Created.Unix()
		}
		return time
	}

	sort.Slice(sources, func(a, b int) bool {
		jobTimeA := extractJobTime(sources[a])
		jobTimeB := extractJobTime(sources[b])
		currentStorageTimeA := extractCurrentStorageTime(sources[a])
		currentStorageTimeB := extractCurrentStorageTime(sources[b])
		if jobTimeA > jobTimeB {
			return true
		}
		if jobTimeA < jobTimeB {
			return false
		}
		return currentStorageTimeA > currentStorageTimeB
	})

	summaries := make([]*userPb.CidSummary, len(sources))
	for i, source := range sources {
		summaries[i] = &userPb.CidSummary{
			Cid:    source.cid,
			Stored: source.currentStorage != nil,
		}
		if source.executingJob != nil {
			summaries[i].ExecutingJob = source.executingJob.ID.String()
		}
		if len(source.queuedJobs) > 0 {
			queuedIds := make([]string, len(source.queuedJobs))
			for j, job := range source.queuedJobs {
				queuedIds[j] = job.ID.String()
			}
			summaries[i].QueuedJobs = queuedIds
		}
	}

	return &userPb.CidSummaryResponse{CidSummary: summaries}, nil
}

// CidInfo returns information about cids managed by the FFS instance.
func (s *Service) CidInfo(ctx context.Context, req *userPb.CidInfoRequest) (*userPb.CidInfoResponse, error) {
	i, err := s.getInstanceByToken(ctx)
	if err != nil {
		return nil, err
	}

	cid, err := fromProtoCid(req.Cid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parsing cid: %v", err)
	}

	storageConfigs, err := i.GetStorageConfigs(cid)
	if err != nil {
		code := codes.Internal
		if err == api.ErrNotFound {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "getting storage configs: %v", err)
	}
	config, ok := storageConfigs[cid]
	if !ok {
		log.Warnf("didn't find storage config for cid %s", cid.String())
		return nil, status.Errorf(codes.Internal, "didn't find storage config for cid %s", cid.String())
	}

	rpcConfig := toRPCStorageConfig(config)
	cidInfo := &userPb.CidInfo{
		Cid:                       cid.String(),
		LatestPushedStorageConfig: rpcConfig,
	}
	info, err := i.StorageInfo(cid)
	if err != nil && err != api.ErrNotFound {
		return nil, status.Errorf(codes.Internal, "getting storage info: %v", err)
	} else if err == nil {
		cidInfo.CurrentStorageInfo = su.ToRPCStorageInfo(info)
	}
	queuedJobs := i.QueuedStorageJobs(cid)
	rpcQueudJobs := make([]*userPb.StorageJob, len(queuedJobs))
	for i, job := range queuedJobs {
		rpcJob, err := su.ToRPCJob(job)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "converting job to rpc job: %v", err)
		}
		rpcQueudJobs[i] = rpcJob
	}
	cidInfo.QueuedStorageJobs = rpcQueudJobs
	executingJobs := i.ExecutingStorageJobs(cid)
	if len(executingJobs) == 1 {
		// There is exactly one job in the slice because we specified a cid
		// and there can be only one executing job per cid at a time.
		rpcJob, err := su.ToRPCJob(executingJobs[0])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "converting job to rpc job: %v", err)
		}
		cidInfo.ExecutingStorageJob = rpcJob
	} else if len(executingJobs) > 1 {
		log.Warnf("received %d executing jobs when there should be 1", len(executingJobs))
	}
	finalJobs := i.LatestFinalStorageJobs(cid)
	if len(finalJobs) > 0 {
		rpcJob, err := su.ToRPCJob(finalJobs[0])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "converting job to rpc job: %v", err)
		}
		cidInfo.LatestFinalStorageJob = rpcJob
	}
	successfulJobs := i.LatestSuccessfulStorageJobs(cid)
	if len(successfulJobs) > 0 {
		rpcJob, err := su.ToRPCJob(successfulJobs[0])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "converting job to rpc job: %v", err)
		}
		cidInfo.LatestSuccessfulStorageJob = rpcJob
	}
	return &userPb.CidInfoResponse{CidInfo: cidInfo}, nil
}

func receiveFile(srv userPb.UserService_StageServer, writer *io.PipeWriter) {
	for {
		req, err := srv.Recv()
		if err == io.EOF {
			_ = writer.Close()
			break
		} else if err != nil {
			_ = writer.CloseWithError(err)
			break
		}
		_, writeErr := writer.Write(req.GetChunk())
		if writeErr != nil {
			if err := writer.CloseWithError(writeErr); err != nil {
				log.Errorf("closing with error: %s", err)
			}
		}
	}
}
