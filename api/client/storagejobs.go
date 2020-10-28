package client

import (
	"context"
	"io"

	proto "github.com/textileio/powergate/proto/powergate/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StorageJobs provides access to Powergate jobs APIs.
type StorageJobs struct {
	client proto.PowergateServiceClient
}

// WatchStorageJobsEvent represents an event for Watching a job.
type WatchStorageJobsEvent struct {
	Res *proto.WatchStorageJobsResponse
	Err error
}

// StorageJob returns the current state of the specified job.
func (j *StorageJobs) StorageJob(ctx context.Context, jid string) (*proto.StorageJobResponse, error) {
	return j.client.StorageJob(ctx, &proto.StorageJobRequest{Jid: jid})
}

// StorageConfigForJob returns the StorageConfig associated with the specified job.
func (j *StorageJobs) StorageConfigForJob(ctx context.Context, jobID string) (*proto.StorageConfigForJobResponse, error) {
	return j.client.StorageConfigForJob(ctx, &proto.StorageConfigForJobRequest{JobId: jobID})
}

// Queued returns a list of queued storage jobs.
func (j *StorageJobs) Queued(ctx context.Context, cids ...string) (*proto.QueuedStorageJobsResponse, error) {
	req := &proto.QueuedStorageJobsRequest{
		Cids: cids,
	}
	return j.client.QueuedStorageJobs(ctx, req)
}

// Executing returns a list of executing storage jobs.
func (j *StorageJobs) Executing(ctx context.Context, cids ...string) (*proto.ExecutingStorageJobsResponse, error) {
	req := &proto.ExecutingStorageJobsRequest{
		Cids: cids,
	}
	return j.client.ExecutingStorageJobs(ctx, req)
}

// LatestFinal returns a list of latest final storage jobs.
func (j *StorageJobs) LatestFinal(ctx context.Context, cids ...string) (*proto.LatestFinalStorageJobsResponse, error) {
	req := &proto.LatestFinalStorageJobsRequest{
		Cids: cids,
	}
	return j.client.LatestFinalStorageJobs(ctx, req)
}

// LatestSuccessful returns a list of latest successful storage jobs.
func (j *StorageJobs) LatestSuccessful(ctx context.Context, cids ...string) (*proto.LatestSuccessfulStorageJobsResponse, error) {
	req := &proto.LatestSuccessfulStorageJobsRequest{
		Cids: cids,
	}
	return j.client.LatestSuccessfulStorageJobs(ctx, req)
}

// Summary returns a summary of storage jobs.
func (j *StorageJobs) Summary(ctx context.Context, cids ...string) (*proto.StorageJobsSummaryResponse, error) {
	req := &proto.StorageJobsSummaryRequest{
		Cids: cids,
	}
	return j.client.StorageJobsSummary(ctx, req)
}

// Watch pushes JobEvents to the provided channel. The provided channel will be owned
// by the client after the call, so it shouldn't be closed by the client. To stop receiving
// events, the provided ctx should be canceled. If an error occurs, it will be returned
// in the Err field of JobEvent and the channel will be closed.
func (j *StorageJobs) Watch(ctx context.Context, ch chan<- WatchStorageJobsEvent, jids ...string) error {
	stream, err := j.client.WatchStorageJobs(ctx, &proto.WatchStorageJobsRequest{Jids: jids})
	if err != nil {
		return err
	}
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- WatchStorageJobsEvent{Err: err}
				close(ch)
				break
			}
			ch <- WatchStorageJobsEvent{Res: res}
		}
	}()
	return nil
}

// Cancel signals that the executing Job with JobID jid should be
// canceled.
func (j *StorageJobs) Cancel(ctx context.Context, jid string) (*proto.CancelStorageJobResponse, error) {
	return j.client.CancelStorageJob(ctx, &proto.CancelStorageJobRequest{Jid: jid})
}
