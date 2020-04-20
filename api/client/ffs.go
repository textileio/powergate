package client

import (
	"context"
	"io"
	"time"

	cid "github.com/ipfs/go-cid"
	ff "github.com/textileio/powergate/ffs"
	"github.com/textileio/powergate/ffs/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ffs struct {
	client rpc.FFSClient
}

// JobEvent represents an event for Watching a job
type JobEvent struct {
	Job ff.Job
	Err error
}

// PushConfigOption mutates a push configuration.
type PushConfigOption func(o *PushConfig)

// PushConfig contains options for pushing a Cid configuration.
type PushConfig struct {
	Config            ff.CidConfig
	HasConfig         bool
	OverrideConfig    bool
	HasOverrideConfig bool
}

// WithCidConfig overrides the Api default Cid configuration.
func WithCidConfig(c ff.CidConfig) PushConfigOption {
	return func(o *PushConfig) {
		o.Config = c
		o.HasConfig = true
	}
}

// WithOverride allows a new push configuration to override an existing one.
// It's used as an extra security measure to avoid unwanted configuration changes.
func WithOverride(override bool) PushConfigOption {
	return func(o *PushConfig) {
		o.OverrideConfig = override
		o.HasOverrideConfig = true
	}
}

// WatchLogsConfig contains configuration for a stream-log
// of human-friendly messages for a Cid execution.
type WatchLogsConfig struct {
	jid ff.JobID
}

// WatchLogsOption is a function that changes GetLogsConfig.
type WatchLogsOption func(config *WatchLogsConfig)

// WithJidFilter filters only log messages of a Cid related to
// the Job with id jid.
func WithJidFilter(jid ff.JobID) WatchLogsOption {
	return func(c *WatchLogsConfig) {
		c.jid = jid
	}
}

// LogEvent represents an event for watching cid logs
type LogEvent struct {
	LogEntry ff.LogEntry
	Err      error
}

func (f *ffs) Create(ctx context.Context) (string, string, error) {
	r, err := f.client.Create(ctx, &rpc.CreateRequest{})
	if err != nil {
		return "", "", err
	}
	return r.ID, r.Token, nil
}

func (f *ffs) ID(ctx context.Context) (ff.APIID, error) {
	resp, err := f.client.ID(ctx, &rpc.IDRequest{})
	if err != nil {
		return ff.EmptyInstanceID, err
	}
	return ff.APIID(resp.ID), nil
}

func (f *ffs) WalletAddr(ctx context.Context) (string, error) {
	resp, err := f.client.WalletAddr(ctx, &rpc.WalletAddrRequest{})
	if err != nil {
		return "", err
	}
	return resp.Addr, nil
}

func (f *ffs) GetDefaultCidConfig(ctx context.Context, c cid.Cid) (*rpc.GetDefaultCidConfigReply, error) {
	return f.client.GetDefaultCidConfig(ctx, &rpc.GetDefaultCidConfigRequest{Cid: c.String()})
}

func (f *ffs) GetCidConfig(ctx context.Context, c cid.Cid) (*rpc.GetCidConfigReply, error) {
	return f.client.GetCidConfig(ctx, &rpc.GetCidConfigRequest{Cid: c.String()})
}

func (f *ffs) SetDefaultCidConfig(ctx context.Context, config ff.DefaultCidConfig) error {
	req := &rpc.SetDefaultCidConfigRequest{
		Config: &rpc.DefaultCidConfig{
			Hot: &rpc.HotConfig{
				Enabled:       config.Hot.Enabled,
				AllowUnfreeze: config.Hot.AllowUnfreeze,
				Ipfs: &rpc.IpfsConfig{
					AddTimeout: int64(config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &rpc.ColdConfig{
				Enabled: config.Cold.Enabled,
				Filecoin: &rpc.FilConfig{
					RepFactor:      int64(config.Cold.Filecoin.RepFactor),
					DealDuration:   int64(config.Cold.Filecoin.DealDuration),
					ExcludedMiners: config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   config.Cold.Filecoin.CountryCodes,
					Renew: &rpc.FilRenew{
						Enabled:   config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		},
	}
	_, err := f.client.SetDefaultCidConfig(ctx, req)
	return err
}

func (f *ffs) Show(ctx context.Context, c cid.Cid) (*rpc.ShowReply, error) {
	return f.client.Show(ctx, &rpc.ShowRequest{
		Cid: c.String(),
	})
}

func (f *ffs) Info(ctx context.Context) (*rpc.InfoReply, error) {
	return f.client.Info(ctx, &rpc.InfoRequest{})
}

func (f *ffs) WatchJobs(ctx context.Context, ch chan<- JobEvent, jids ...ff.JobID) error {
	jidStrings := make([]string, len(jids))
	for i, jid := range jids {
		jidStrings[i] = jid.String()
	}

	stream, err := f.client.WatchJobs(ctx, &rpc.WatchJobsRequest{Jids: jidStrings})
	if err != nil {
		return err
	}
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- JobEvent{Err: err}
				close(ch)
				break
			}
			job := ff.Job{
				ID:       ff.JobID(reply.Job.ID),
				APIID:    ff.APIID(reply.Job.ApiID),
				Status:   ff.JobStatus(reply.Job.Status),
				ErrCause: reply.Job.ErrCause,
			}
			ch <- JobEvent{Job: job}
		}
	}()
	return nil
}

func (f *ffs) Replace(ctx context.Context, c1 cid.Cid, c2 cid.Cid) (ff.JobID, error) {
	resp, err := f.client.Replace(ctx, &rpc.ReplaceRequest{Cid1: c1.String(), Cid2: c2.String()})
	if err != nil {
		return ff.EmptyJobID, err
	}
	return ff.JobID(resp.JobID), nil
}

func (f *ffs) PushConfig(ctx context.Context, c cid.Cid, opts ...PushConfigOption) (ff.JobID, error) {
	pushConfig := PushConfig{}
	for _, opt := range opts {
		opt(&pushConfig)
	}

	req := &rpc.PushConfigRequest{Cid: c.String()}

	if pushConfig.HasConfig {
		req.HasConfig = true
		req.Config = &rpc.CidConfig{
			Cid: pushConfig.Config.Cid.String(),
			Hot: &rpc.HotConfig{
				Enabled:       pushConfig.Config.Hot.Enabled,
				AllowUnfreeze: pushConfig.Config.Hot.AllowUnfreeze,
				Ipfs: &rpc.IpfsConfig{
					AddTimeout: int64(pushConfig.Config.Hot.Ipfs.AddTimeout),
				},
			},
			Cold: &rpc.ColdConfig{
				Enabled: pushConfig.Config.Cold.Enabled,
				Filecoin: &rpc.FilConfig{
					RepFactor:      int64(pushConfig.Config.Cold.Filecoin.RepFactor),
					DealDuration:   pushConfig.Config.Cold.Filecoin.DealDuration,
					ExcludedMiners: pushConfig.Config.Cold.Filecoin.ExcludedMiners,
					CountryCodes:   pushConfig.Config.Cold.Filecoin.CountryCodes,
					Renew: &rpc.FilRenew{
						Enabled:   pushConfig.Config.Cold.Filecoin.Renew.Enabled,
						Threshold: int64(pushConfig.Config.Cold.Filecoin.Renew.Threshold),
					},
				},
			},
		}
	}

	if pushConfig.HasOverrideConfig {
		req.HasOverrideConfig = true
		req.OverrideConfig = pushConfig.OverrideConfig
	}

	resp, err := f.client.PushConfig(ctx, req)
	if err != nil {
		return ff.EmptyJobID, err
	}

	return ff.JobID(resp.JobID), nil
}

func (f *ffs) Remove(ctx context.Context, c cid.Cid) error {
	_, err := f.client.Remove(ctx, &rpc.RemoveRequest{Cid: c.String()})
	return err
}

func (f *ffs) Get(ctx context.Context, c cid.Cid) (io.Reader, error) {
	stream, err := f.client.Get(ctx, &rpc.GetRequest{
		Cid: c.String(),
	})
	if err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				_ = writer.Close()
				break
			} else if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
			_, err = writer.Write(reply.GetChunk())
			if err != nil {
				_ = writer.CloseWithError(err)
				break
			}
		}
	}()

	return reader, nil
}

func (f *ffs) WatchLogs(ctx context.Context, ch chan<- LogEvent, c cid.Cid, opts ...WatchLogsOption) error {
	config := WatchLogsConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	stream, err := f.client.WatchLogs(ctx, &rpc.WatchLogsRequest{Cid: c.String(), Jid: config.jid.String()})
	if err != nil {
		return err
	}
	go func() {
		for {
			reply, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				close(ch)
				break
			}
			if err != nil {
				ch <- LogEvent{Err: err}
				close(ch)
				break
			}

			cid, err := cid.Decode(reply.LogEntry.Cid)
			if err != nil {
				ch <- LogEvent{Err: err}
				close(ch)
				break
			}

			entry := ff.LogEntry{
				Cid:       cid,
				Timestamp: time.Unix(reply.LogEntry.Time, 0),
				Jid:       ff.JobID(reply.LogEntry.Jid),
				Msg:       reply.LogEntry.Msg,
			}
			ch <- LogEvent{LogEntry: entry}
		}
	}()
	return nil
}

func (f *ffs) Close(ctx context.Context) error {
	_, err := f.client.Close(ctx, &rpc.CloseRequest{})
	return err
}

func (f *ffs) AddToHot(ctx context.Context, data io.Reader) (*cid.Cid, error) {
	stream, err := f.client.AddToHot(ctx)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024*32) // 32KB
	for {
		bytesRead, err := data.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		sendErr := stream.Send(&rpc.AddToHotRequest{Chunk: buffer[:bytesRead]})
		if sendErr != nil {
			if sendErr == io.EOF {
				var noOp interface{}
				return nil, stream.RecvMsg(noOp)
			}
			return nil, sendErr
		}
		if err == io.EOF {
			break
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	cid, err := cid.Decode(reply.GetCid())
	if err != nil {
		return nil, err
	}
	return &cid, nil
}
